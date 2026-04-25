package outbox

import "time"

type OutboxRecord struct {
	ID          string
	EventName   string
	Payload     []byte
	OccurredAt  time.Time
	ProcessedAt *time.Time
}
