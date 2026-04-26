package usecases

import (
	"context"
	"fmt"

	"ddd-core-banking/internal/notification/infrastructure/email"
)

type NotifyRejectionUseCase struct {
	sender email.Sender
}

func NewNotifyRejectionUseCase(sender email.Sender) *NotifyRejectionUseCase {
	return &NotifyRejectionUseCase{sender: sender}
}

type NotifyRejectionInput struct {
	FullName string
	Email    string
	Reason   string
}

func (uc *NotifyRejectionUseCase) Execute(ctx context.Context, input NotifyRejectionInput) error {
	subject := "Sua solicitação de conta foi reprovada"
	body := fmt.Sprintf(
		"Olá, %s.\n\nSua solicitação foi reprovada pelo seguinte motivo: %s\n\nVocê poderá tentar novamente em 90 dias.",
		input.FullName,
		input.Reason,
	)
	if err := uc.sender.Send(ctx, input.Email, subject, body); err != nil {
		return fmt.Errorf("sending rejection email to %s: %w", input.Email, err)
	}
	return nil
}
