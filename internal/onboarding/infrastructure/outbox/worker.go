package outbox

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"ddd-core-banking/internal/onboarding/domain/repository"
)

// MessagePublisher envia o payload bruto (JSON) para o broker (Kafka, RabbitMQ, etc.).
// Separado do EventBus porque o worker opera sobre bytes serializados, não DomainEvent tipado.
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

// Start bloqueia a goroutine atual, ouvindo o canal 'outbox_event' do Postgres.
// Acorda automaticamente quando o Save faz pg_notify após o commit.
// Encerra sem erro quando ctx é cancelado.
func (w *Worker) Start(ctx context.Context) error {
	if _, err := w.conn.Exec(ctx, "LISTEN outbox_event"); err != nil {
		return fmt.Errorf("subscribing to outbox channel: %w", err)
	}

	// processa registros que ficaram pendentes antes do worker iniciar
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
