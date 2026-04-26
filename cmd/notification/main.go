package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ddd-core-banking/internal/notification/application/usecases"
	"ddd-core-banking/internal/notification/infrastructure/email"
	"ddd-core-banking/internal/notification/interfaces/eventhandler"
	"ddd-core-banking/pkg/messaging/rabbitmq"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		log.Fatal("RABBITMQ_URL is required")
	}

	consumer, err := rabbitmq.NewConsumer(rabbitURL)
	if err != nil {
		log.Fatalf("creating rabbitmq consumer: %v", err)
	}
	defer consumer.Close()

	emailSender := email.NewStubSender()
	notifyRejection := usecases.NewNotifyRejectionUseCase(emailSender)
	clientRejected := eventhandler.NewClientRejectedHandler(notifyRejection)

	if err := consumer.Subscribe("core-banking", "notification", "Onboarding.ClientRejected", clientRejected.Handle); err != nil {
		log.Fatalf("subscribing to Onboarding.ClientRejected: %v", err)
	}

	log.Println("notification module listening for events...")
	<-ctx.Done()
	log.Println("shutting down...")
}
