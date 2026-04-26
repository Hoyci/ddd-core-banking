package usecases

import (
	"fmt"

	"ddd-core-banking/internal/core-banking/domain/repository"
)

type DebitAccountUseCase struct {
	repo repository.AccountRepository
}

func NewDebitAccountUseCase(repo repository.AccountRepository) *DebitAccountUseCase {
	return &DebitAccountUseCase{repo: repo}
}

type DebitAccountInput struct {
	AccountID string
	Amount    int64
}

func (uc *DebitAccountUseCase) Execute(input DebitAccountInput) error {
	if err := uc.repo.Debit(input.AccountID, input.Amount); err != nil {
		return fmt.Errorf("debiting account: %w", err)
	}
	return nil
}
