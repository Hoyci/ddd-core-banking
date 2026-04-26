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

type TransferFundsUseCase struct {
	repo        repository.PaymentRepository
	coreBanking corebanking.Client
}

func NewTransferFundsUseCase(repo repository.PaymentRepository, cb corebanking.Client) *TransferFundsUseCase {
	return &TransferFundsUseCase{repo: repo, coreBanking: cb}
}

type TransferFundsInput struct {
	SenderAccountID   string
	ReceiverAccountID string
	Amount            int64
}

type TransferFundsOutput struct {
	TransferID string `json:"transfer_id"`
	Amount     int64  `json:"amount"`
}

func (uc *TransferFundsUseCase) Execute(ctx context.Context, input TransferFundsInput) (TransferFundsOutput, error) {
	sender, err := uc.repo.GetAccountSnapshot(input.SenderAccountID)
	if err != nil {
		return TransferFundsOutput{}, err
	}
	receiver, err := uc.repo.GetAccountSnapshot(input.ReceiverAccountID)
	if err != nil {
		return TransferFundsOutput{}, err
	}

	if sender.Blocked || receiver.Blocked {
		return TransferFundsOutput{}, payerrors.ErrAccountBlocked
	}
	if sender.Balance < input.Amount {
		return TransferFundsOutput{}, payerrors.ErrInsufficientFunds
	}

	transfer, err := entity.CreateTransfer(entity.CreateTransferInput{
		SenderAccountID:   input.SenderAccountID,
		ReceiverAccountID: input.ReceiverAccountID,
		Amount:            input.Amount,
	})
	if err != nil {
		return TransferFundsOutput{}, err
	}

	if err := uc.coreBanking.ProcessTransfer(ctx, corebanking.TransferRequest{
		SenderAccountID:   input.SenderAccountID,
		ReceiverAccountID: input.ReceiverAccountID,
		Amount:            input.Amount,
	}); err != nil {
		return TransferFundsOutput{}, fmt.Errorf("%w: %v", payerrors.ErrCoreBankingUnavailable, err)
	}

	evt := event.NewTransferProcessed(transfer.TransferID(), input.SenderAccountID, input.ReceiverAccountID, transfer.Amount())
	if err := uc.repo.SaveTransfer(transfer, []events.DomainEvent{evt}); err != nil {
		return TransferFundsOutput{}, fmt.Errorf("saving transfer: %w", err)
	}

	return TransferFundsOutput{TransferID: transfer.TransferID(), Amount: transfer.Amount()}, nil
}
