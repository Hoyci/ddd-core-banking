package email

import "context"

type Sender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// StubSender is a no-op implementation for local development.
// Replace with an SMTP or transactional email provider in production.
type StubSender struct{}

func NewStubSender() *StubSender { return &StubSender{} }

func (s *StubSender) Send(_ context.Context, _, _, _ string) error { return nil }
