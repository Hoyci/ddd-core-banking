package event

import "time"

type InvoicePaymentProcessed struct {
	PaymentID string    `json:"payment_id"`
	AccountID string    `json:"account_id"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

func NewInvoicePaymentProcessed(paymentID, accountID string, amount int64) InvoicePaymentProcessed {
	return InvoicePaymentProcessed{
		PaymentID: paymentID,
		AccountID: accountID,
		Amount:    amount,
		CreatedAt: time.Now(),
	}
}

func (e InvoicePaymentProcessed) EventName() string     { return "Payment.InvoicePaymentProcessed" }
func (e InvoicePaymentProcessed) OccurredAt() time.Time { return e.CreatedAt }
