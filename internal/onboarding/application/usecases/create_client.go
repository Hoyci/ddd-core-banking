package usecases

import (
	"ddd-core-banking/internal/onboarding/domain"
	"ddd-core-banking/internal/onboarding/domain/entity"
	"ddd-core-banking/internal/onboarding/domain/repository"
	"errors"
	"fmt"
)

type CreateClientUseCase struct {
	repo repository.ClientRepository
}

func NewCreateClientUseCase(repo repository.ClientRepository) *CreateClientUseCase {
	return &CreateClientUseCase{repo: repo}
}

func (uc *CreateClientUseCase) Execute(input entity.CreateClientInput) error {
	existing, err := uc.repo.FindByEmail(input.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("finding client by email: %w", err)
	}
	if existing != nil {
		return domain.ErrEmailAlreadyInUse
	}

	client, err := entity.CreateClient(input)
	if err != nil {
		return err
	}

	return uc.repo.Save(client, client.PullEvents())
}
