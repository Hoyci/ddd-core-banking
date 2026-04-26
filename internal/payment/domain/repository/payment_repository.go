package repository

import (
	"ddd-core-banking/internal/payment/domain/entity"
	"ddd-core-banking/pkg/events"
)

type AccountSnapshot struct {
	Balance int64
	Blocked bool
}

type PaymentRepository interface {
	GetAccountSnapshot(accountID string) (AccountSnapshot, error)
	SaveInvoicePayment(payment *entity.InvoicePayment, domainEvents []events.DomainEvent) error
	SaveTransfer(transfer *entity.Transfer, domainEvents []events.DomainEvent) error
}
