package corebanking

import "context"

type InvoicePaymentRequest struct {
	Barcode   string
	AccountID string
	Amount    int64
}

type TransferRequest struct {
	SenderAccountID   string
	ReceiverAccountID string
	Amount            int64
}

type Client interface {
	ProcessInvoicePayment(ctx context.Context, req InvoicePaymentRequest) error
	ProcessTransfer(ctx context.Context, req TransferRequest) error
}

// StubClient satisfies the Client interface with no-op calls for local development.
// Replace with an HTTPClient pointing to the real core banking service in production.
type StubClient struct{}

func NewStubClient() *StubClient { return &StubClient{} }

func (c *StubClient) ProcessInvoicePayment(_ context.Context, _ InvoicePaymentRequest) error {
	return nil
}

func (c *StubClient) ProcessTransfer(_ context.Context, _ TransferRequest) error {
	return nil
}
