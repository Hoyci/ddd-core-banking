package events

import "time"

// DomainEvent é a interface que todo evento de domínio deve implementar.
// Qualquer struct em domain/event/ de qualquer bounded context assina esse contrato.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// EventBus é o contrato de publicação.
// A implementação concreta (Kafka, RabbitMQ, in-memory) fica em infrastructure/.
// O domínio e a application só enxergam essa interface.
type EventBus interface {
	Publish(events ...DomainEvent) error
}
