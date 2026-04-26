package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	paymentusecases "ddd-core-banking/internal/payment/application/usecases"
	"ddd-core-banking/internal/payment/infrastructure/corebanking"
	"ddd-core-banking/internal/payment/infrastructure/outbox"
	paymentpostgres "ddd-core-banking/internal/payment/infrastructure/persistence/postgres"
	paymenthandler "ddd-core-banking/internal/payment/interfaces/http"
	"ddd-core-banking/pkg/messaging/rabbitmq"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		log.Fatal("RABBITMQ_URL is required")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connecting to database pool: %v", err)
	}
	defer pool.Close()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("connecting to database for outbox worker: %v", err)
	}
	defer conn.Close(ctx)

	publisher, err := rabbitmq.NewPublisher(rabbitURL, "core-banking")
	if err != nil {
		log.Fatalf("creating rabbitmq publisher: %v", err)
	}
	defer publisher.Close()

	paymentRepo := paymentpostgres.NewPaymentRepository(pool)
	outboxRepo := paymentpostgres.NewOutboxRepository(pool)
	cbClient := corebanking.NewStubClient()

	payInvoice := paymentusecases.NewPayInvoiceUseCase(paymentRepo, cbClient)
	transferFunds := paymentusecases.NewTransferFundsUseCase(paymentRepo, cbClient)

	handler := paymenthandler.NewPaymentHandler(payInvoice, transferFunds)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /payments/invoice", handler.PayInvoice)
	mux.HandleFunc("POST /payments/transfers", handler.TransferFunds)

	workerErr := make(chan error, 1)
	go func() {
		worker := outbox.NewWorker(conn, outboxRepo, publisher)
		workerErr <- worker.Start(ctx)
	}()

	server := &http.Server{Addr: ":8081", Handler: mux}
	serverErr := make(chan error, 1)
	go func() {
		log.Println("payment server listening on :8081")
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down...")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("http server shutdown error: %v", err)
		}
	case err := <-workerErr:
		log.Fatalf("outbox worker stopped: %v", err)
	case err := <-serverErr:
		log.Fatalf("http server stopped: %v", err)
	}
}
