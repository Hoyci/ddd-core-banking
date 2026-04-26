package repository

import "ddd-core-banking/pkg/outbox"

type OutboxRepository interface {
	FindPending() ([]outbox.OutboxRecord, error)
	MarkAsProcessed(id string) error
}
