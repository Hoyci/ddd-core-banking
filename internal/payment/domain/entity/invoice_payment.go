package entity

import (
	"time"

	payerrors "ddd-core-banking/internal/payment/domain/errors"
	"ddd-core-banking/pkg/valueobjects"
)

type InvoicePayment struct {
	paymentID string
	accountID string
	barcode   string
	amount    int64
	createdAt time.Time
}

type CreateInvoicePaymentInput struct {
	AccountID string
	Barcode   string
	Amount    int64
}

func CreateInvoicePayment(input CreateInvoicePaymentInput) (*InvoicePayment, error) {
	if input.AccountID == "" {
		return nil, payerrors.ErrAccountIDRequired
	}
	if input.Barcode == "" {
		return nil, payerrors.ErrBarcodeRequired
	}
	if input.Amount <= 0 {
		return nil, payerrors.ErrAmountInvalid
	}
	return &InvoicePayment{
		paymentID: valueobjects.GenerateID(),
		accountID: input.AccountID,
		barcode:   input.Barcode,
		amount:    input.Amount,
		createdAt: time.Now(),
	}, nil
}

func (b *InvoicePayment) PaymentID() string    { return b.paymentID }
func (b *InvoicePayment) AccountID() string    { return b.accountID }
func (b *InvoicePayment) Barcode() string      { return b.barcode }
func (b *InvoicePayment) Amount() int64        { return b.amount }
func (b *InvoicePayment) CreatedAt() time.Time { return b.createdAt }
