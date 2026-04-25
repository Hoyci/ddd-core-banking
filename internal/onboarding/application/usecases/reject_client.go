package usecases

import (
	"ddd-core-banking/internal/onboarding/domain/repository"
	"fmt"
)

type RejectClientUseCase struct {
	repo repository.ClientRepository
}

func NewRejectClientUseCase(repo repository.ClientRepository) *RejectClientUseCase {
	return &RejectClientUseCase{
		repo: repo,
	}
}

type RejectClientInput struct {
	ClientID string
	Reason   string
}

func (uc *RejectClientUseCase) Execute(input RejectClientInput) error {
	client, err := uc.repo.FindByID(input.ClientID)
	if err != nil {
		return fmt.Errorf("finding client by id: %w", err)
	}

	if err := client.RejectClient(input.Reason); err != nil {
		return fmt.Errorf("rejecting client: %w", err)
	}

	return uc.repo.Save(client, client.PullEvents())
}
