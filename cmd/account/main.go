package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	accountusecases "ddd-core-banking/internal/account/application/usecases"
	accountpostgres "ddd-core-banking/internal/account/infrastructure/persistence/postgres"
	accounteventhandler "ddd-core-banking/internal/account/interfaces/eventhandler"
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

	accountRepo := accountpostgres.NewAccountRepository(pool)
	accountNumberGen := accountpostgres.NewSequenceAccountNumberGenerator(pool)
	createAccount := accountusecases.NewCreateAccountUseCase(accountRepo, accountNumberGen)

	consumer, err := rabbitmq.NewConsumer(rabbitURL)
	if err != nil {
		log.Fatalf("creating rabbitmq consumer: %v", err)
	}
	defer consumer.Close()

	clientApproved := accounteventhandler.NewClientApprovedHandler(createAccount)
	if err := consumer.Subscribe("core-banking", "account", "Onboarding.ClientApproved", clientApproved.Handle); err != nil {
		log.Fatalf("subscribing to Onboarding.ClientApproved: %v", err)
	}

	log.Println("account module listening for events...")
	<-ctx.Done()
	log.Println("shutting down...")
}
