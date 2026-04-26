package usecases

import (
	"fmt"

	"ddd-core-banking/internal/core-banking/domain/entity"
	"ddd-core-banking/internal/core-banking/domain/repository"
	"ddd-core-banking/internal/core-banking/domain/service"
)

type CreateAccountUseCase struct {
	repo      repository.AccountRepository
	generator service.AccountNumberGenerator
}

func NewCreateAccountUseCase(repo repository.AccountRepository, generator service.AccountNumberGenerator) *CreateAccountUseCase {
	return &CreateAccountUseCase{repo: repo, generator: generator}
}

type CreateAccountInput struct {
	ClientID string
}

func (uc *CreateAccountUseCase) Execute(input CreateAccountInput) error {
	number, err := uc.generator.Next()
	if err != nil {
		return fmt.Errorf("generating account number: %w", err)
	}

	account, err := entity.CreateAccount(entity.CreateAccountInput{
		ClientID: input.ClientID,
		Number:   number,
	})
	if err != nil {
		return err
	}

	if err := uc.repo.Save(account); err != nil {
		return fmt.Errorf("saving account: %w", err)
	}

	return nil
}
