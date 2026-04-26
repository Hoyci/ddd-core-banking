package usecases

import (
	"context"
	"fmt"

	"ddd-core-banking/internal/payment/domain/entity"
	payerrors "ddd-core-banking/internal/payment/domain/errors"
	"ddd-core-banking/internal/payment/domain/event"
	"ddd-core-banking/internal/payment/domain/repository"
	"ddd-core-banking/internal/payment/infrastructure/corebanking"
	"ddd-core-banking/pkg/events"
)

type PayInvoiceUseCase struct {
	repo        repository.PaymentRepository
	coreBanking corebanking.Client
}

func NewPayInvoiceUseCase(repo repository.PaymentRepository, cb corebanking.Client) *PayInvoiceUseCase {
	return &PayInvoiceUseCase{repo: repo, coreBanking: cb}
}

type PayInvoiceInput struct {
	AccountID string
	Barcode   string
	Amount    int64
}

type PayInvoiceOutput struct {
	PaymentID string `json:"payment_id"`
	Amount    int64  `json:"amount"`
}

func (uc *PayInvoiceUseCase) Execute(ctx context.Context, input PayInvoiceInput) (PayInvoiceOutput, error) {
	snapshot, err := uc.repo.GetAccountSnapshot(input.AccountID)
	if err != nil {
		return PayInvoiceOutput{}, err
	}
	if snapshot.Blocked {
		return PayInvoiceOutput{}, payerrors.ErrAccountBlocked
	}
	if snapshot.Balance < input.Amount {
		return PayInvoiceOutput{}, payerrors.ErrInsufficientFunds
	}

	payment, err := entity.CreateInvoicePayment(entity.CreateInvoicePaymentInput{
		AccountID: input.AccountID,
		Barcode:   input.Barcode,
		Amount:    input.Amount,
	})
	if err != nil {
		return PayInvoiceOutput{}, err
	}

	if err := uc.coreBanking.ProcessInvoicePayment(ctx, corebanking.InvoicePaymentRequest{
		Barcode:   input.Barcode,
		AccountID: input.AccountID,
		Amount:    input.Amount,
	}); err != nil {
		return PayInvoiceOutput{}, fmt.Errorf("%w: %v", payerrors.ErrCoreBankingUnavailable, err)
	}

	evt := event.NewInvoicePaymentProcessed(payment.PaymentID(), payment.AccountID(), payment.Amount())
	if err := uc.repo.SaveInvoicePayment(payment, []events.DomainEvent{evt}); err != nil {
		return PayInvoiceOutput{}, fmt.Errorf("saving invoice payment: %w", err)
	}

	return PayInvoiceOutput{PaymentID: payment.PaymentID(), Amount: payment.Amount()}, nil
}
