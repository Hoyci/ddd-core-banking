package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	corebankingusecases "ddd-core-banking/internal/core-banking/application/usecases"
	corebankingpostgres "ddd-core-banking/internal/core-banking/infrastructure/persistence/postgres"
	corebankingeventhandler "ddd-core-banking/internal/core-banking/interfaces/eventhandler"
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
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	accountRepo := corebankingpostgres.NewAccountRepository(pool)
	accountNumberGen := corebankingpostgres.NewSequenceAccountNumberGenerator(pool)

	createAccount := corebankingusecases.NewCreateAccountUseCase(accountRepo, accountNumberGen)
	debitAccount := corebankingusecases.NewDebitAccountUseCase(accountRepo)
	transferBalance := corebankingusecases.NewTransferBalanceUseCase(accountRepo)

	consumer, err := rabbitmq.NewConsumer(rabbitURL)
	if err != nil {
		log.Fatalf("creating rabbitmq consumer: %v", err)
	}
	defer consumer.Close()

	clientApproved := corebankingeventhandler.NewClientApprovedHandler(createAccount)
	invoiceProcessed := corebankingeventhandler.NewInvoicePaymentProcessedHandler(debitAccount)
	transferProcessed := corebankingeventhandler.NewTransferProcessedHandler(transferBalance)

	if err := consumer.Subscribe("core-banking", "account", "Onboarding.ClientApproved", clientApproved.Handle); err != nil {
		log.Fatalf("subscribing to Onboarding.ClientApproved: %v", err)
	}
	if err := consumer.Subscribe("core-banking", "account-invoice", "Payment.InvoicePaymentProcessed", invoiceProcessed.Handle); err != nil {
		log.Fatalf("subscribing to Payment.InvoicePaymentProcessed: %v", err)
	}
	if err := consumer.Subscribe("core-banking", "account-transfer", "Payment.TransferProcessed", transferProcessed.Handle); err != nil {
		log.Fatalf("subscribing to Payment.TransferProcessed: %v", err)
	}

	log.Println("account module listening for events...")
	<-ctx.Done()
	log.Println("shutting down...")
}
