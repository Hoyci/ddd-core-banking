package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"ddd-core-banking/pkg/outbox"
)

type OutboxRepository struct {
	db *pgxpool.Pool
}

func NewOutboxRepository(db *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) FindPending() ([]outbox.OutboxRecord, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, event_name, payload, occurred_at
		FROM payment_outbox
		WHERE processed_at IS NULL
		ORDER BY occurred_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying pending payment outbox records: %w", err)
	}
	defer rows.Close()

	var records []outbox.OutboxRecord
	for rows.Next() {
		var record outbox.OutboxRecord
		if err := rows.Scan(&record.ID, &record.EventName, &record.Payload, &record.OccurredAt); err != nil {
			return nil, fmt.Errorf("scanning payment outbox record: %w", err)
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *OutboxRepository) MarkAsProcessed(id string) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE payment_outbox SET processed_at = $1 WHERE id = $2
	`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("marking payment outbox record %s as processed: %w", id, err)
	}
	return nil
}
