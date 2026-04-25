package repository

import (
	"ddd-core-banking/internal/onboarding/domain/entity"
	"ddd-core-banking/pkg/events"
)

type ClientRepository interface {
	Save(client *entity.Client, events []events.DomainEvent) error
	FindByID(id string) (*entity.Client, error)
	FindByEmail(email string) (*entity.Client, error)
}
