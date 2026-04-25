package usecases

import (
	"ddd-core-banking/internal/onboarding/domain/repository"
	"fmt"
)

type ApproveClientUseCase struct {
	repo repository.ClientRepository
}

func NewApproveClientUseCase(repo repository.ClientRepository) *ApproveClientUseCase {
	return &ApproveClientUseCase{
		repo: repo,
	}
}

type ApproveClientInput struct {
	ClientID string
}

func (uc *ApproveClientUseCase) Execute(input ApproveClientInput) error {
	client, err := uc.repo.FindByID(input.ClientID)
	if err != nil {
		return fmt.Errorf("finding client by id: %w", err)
	}

	if err := client.ApproveClient(); err != nil {
		return fmt.Errorf("approving client: %w", err)
	}

	return uc.repo.Save(client, client.PullEvents())
}
