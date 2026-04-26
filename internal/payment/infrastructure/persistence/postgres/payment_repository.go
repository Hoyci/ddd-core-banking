package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ddd-core-banking/internal/payment/domain/entity"
	payerrors "ddd-core-banking/internal/payment/domain/errors"
	"ddd-core-banking/internal/payment/domain/repository"
	"ddd-core-banking/pkg/events"
	"ddd-core-banking/pkg/valueobjects"
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) GetAccountSnapshot(accountID string) (repository.AccountSnapshot, error) {
	var balance int64
	var blocked bool
	err := r.db.QueryRow(context.Background(), `
		SELECT balance, blocked IS NOT NULL
		FROM accounts WHERE id = $1
	`, accountID).Scan(&balance, &blocked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.AccountSnapshot{}, payerrors.ErrAccountNotFound
		}
		return repository.AccountSnapshot{}, fmt.Errorf("querying account snapshot: %w", err)
	}
	return repository.AccountSnapshot{Balance: balance, Blocked: blocked}, nil
}

func (r *PaymentRepository) SaveInvoicePayment(payment *entity.InvoicePayment, domainEvents []events.DomainEvent) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		INSERT INTO invoice_payments (id, account_id, barcode, amount, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`,
		payment.PaymentID(),
		payment.AccountID(),
		payment.Barcode(),
		payment.Amount(),
		payment.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("inserting invoice payment: %w", err)
	}

	_, err = tx.Exec(context.Background(), `
		INSERT INTO entries (id, account_id, type, amount, source_type, source_id, created_at)
		VALUES ($1, $2, 'DEBIT', $3, 'INVOICE_PAYMENT', $4, $5)
	`,
		valueobjects.GenerateID(),
		payment.AccountID(),
		payment.Amount(),
		payment.PaymentID(),
		payment.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("inserting invoice entry: %w", err)
	}

	if err := insertOutboxEvents(tx, domainEvents); err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (r *PaymentRepository) SaveTransfer(transfer *entity.Transfer, domainEvents []events.DomainEvent) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		INSERT INTO transfers (id, amount, created_at)
		VALUES ($1, $2, $3)
	`,
		transfer.TransferID(),
		transfer.Amount(),
		transfer.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("inserting transfer: %w", err)
	}

	for _, entry := range transfer.Entries() {
		_, err = tx.Exec(context.Background(), `
			INSERT INTO entries (id, account_id, type, amount, source_type, source_id, created_at)
			VALUES ($1, $2, $3, $4, 'TRANSFER', $5, $6)
		`,
			valueobjects.GenerateID(),
			entry.AccountID(),
			entry.EntryType(),
			entry.Amount(),
			transfer.TransferID(),
			transfer.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("inserting transfer entry: %w", err)
		}
	}

	if err := insertOutboxEvents(tx, domainEvents); err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func insertOutboxEvents(tx pgx.Tx, domainEvents []events.DomainEvent) error {
	for _, evt := range domainEvents {
		payload, err := json.Marshal(evt)
		if err != nil {
			return fmt.Errorf("marshaling event %s: %w", evt.EventName(), err)
		}

		_, err = tx.Exec(context.Background(), `
			INSERT INTO payment_outbox (event_name, payload, occurred_at)
			VALUES ($1, $2, $3)
		`,
			evt.EventName(),
			payload,
			evt.OccurredAt(),
		)
		if err != nil {
			return fmt.Errorf("inserting payment outbox record for %s: %w", evt.EventName(), err)
		}
	}

	_, err := tx.Exec(context.Background(), `SELECT pg_notify('payment_outbox_event', '')`)
	if err != nil {
		return fmt.Errorf("notifying payment outbox channel: %w", err)
	}

	return nil
}
