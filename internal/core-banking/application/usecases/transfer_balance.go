package usecases

import (
	"fmt"

	"ddd-core-banking/internal/core-banking/domain/repository"
)

type TransferBalanceUseCase struct {
	repo repository.AccountRepository
}

func NewTransferBalanceUseCase(repo repository.AccountRepository) *TransferBalanceUseCase {
	return &TransferBalanceUseCase{repo: repo}
}

type TransferBalanceInput struct {
	SenderAccountID   string
	ReceiverAccountID string
	Amount            int64
}

func (uc *TransferBalanceUseCase) Execute(input TransferBalanceInput) error {
	if err := uc.repo.TransferBalance(input.SenderAccountID, input.ReceiverAccountID, input.Amount); err != nil {
		return fmt.Errorf("transferring balance: %w", err)
	}
	return nil
}
