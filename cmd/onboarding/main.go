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

	"ddd-core-banking/internal/onboarding/application/usecases"
	"ddd-core-banking/internal/onboarding/infrastructure/outbox"
	"ddd-core-banking/internal/onboarding/infrastructure/persistence/postgres"
	handler "ddd-core-banking/internal/onboarding/interfaces/http"
)

// logPublisher is a stub — replace with a real Kafka/RabbitMQ implementation.
type logPublisher struct{}

func (p *logPublisher) Publish(eventName string, payload []byte) error {
	log.Printf("event published: %s %s", eventName, payload)
	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
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

	clientRepo := postgres.NewClientRepository(pool)
	outboxRepo := postgres.NewOutboxRepository(pool)

	createClient := usecases.NewCreateClientUseCase(clientRepo)
	approveClient := usecases.NewApproveClientUseCase(clientRepo)
	rejectClient := usecases.NewRejectClientUseCase(clientRepo)

	clientHandler := handler.NewClientHandler(createClient, approveClient, rejectClient)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /clients", clientHandler.Create)
	mux.HandleFunc("PATCH /clients/{clientID}/approve", clientHandler.Approve)
	mux.HandleFunc("PATCH /clients/{clientID}/reject", clientHandler.Reject)

	// outbox worker em goroutine separada
	workerErr := make(chan error, 1)
	go func() {
		worker := outbox.NewWorker(conn, outboxRepo, &logPublisher{})
		workerErr <- worker.Start(ctx)
	}()

	// servidor HTTP em goroutine separada
	server := &http.Server{Addr: ":8080", Handler: mux}
	serverErr := make(chan error, 1)
	go func() {
		log.Println("server listening on :8080")
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
