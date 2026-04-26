package outbox

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"ddd-core-banking/internal/payment/domain/repository"
)

type MessagePublisher interface {
	Publish(eventName string, payload []byte) error
}

type Worker struct {
	conn      *pgx.Conn
	outbox    repository.OutboxRepository
	publisher MessagePublisher
}

func NewWorker(conn *pgx.Conn, outbox repository.OutboxRepository, publisher MessagePublisher) *Worker {
	return &Worker{conn: conn, outbox: outbox, publisher: publisher}
}

func (w *Worker) Start(ctx context.Context) error {
	if _, err := w.conn.Exec(ctx, "LISTEN payment_outbox_event"); err != nil {
		return fmt.Errorf("subscribing to payment outbox channel: %w", err)
	}

	if err := w.processAll(); err != nil {
		return err
	}

	for {
		_, err := w.conn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("waiting for notification: %w", err)
		}

		if err := w.processAll(); err != nil {
			return err
		}
	}
}

func (w *Worker) processAll() error {
	records, err := w.outbox.FindPending()
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := w.publisher.Publish(record.EventName, record.Payload); err != nil {
			return err
		}
		if err := w.outbox.MarkAsProcessed(record.ID); err != nil {
			return err
		}
	}

	return nil
}
