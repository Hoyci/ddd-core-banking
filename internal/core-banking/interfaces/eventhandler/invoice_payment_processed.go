package eventhandler

import (
	"encoding/json"
	"fmt"

	"ddd-core-banking/internal/core-banking/application/usecases"
)

type invoicePaymentProcessedPayload struct {
	PaymentID string `json:"payment_id"`
	AccountID string `json:"account_id"`
	Amount    int64  `json:"amount"`
}

type InvoicePaymentProcessedHandler struct {
	useCase *usecases.DebitAccountUseCase
}

func NewInvoicePaymentProcessedHandler(uc *usecases.DebitAccountUseCase) *InvoicePaymentProcessedHandler {
	return &InvoicePaymentProcessedHandler{useCase: uc}
}

func (h *InvoicePaymentProcessedHandler) Handle(payload []byte) error {
	var p invoicePaymentProcessedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshaling InvoicePaymentProcessed payload: %w", err)
	}
	return h.useCase.Execute(usecases.DebitAccountInput{
		AccountID: p.AccountID,
		Amount:    p.Amount,
	})
}
