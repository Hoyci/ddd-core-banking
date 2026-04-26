package eventhandler

import (
	"encoding/json"
	"fmt"

	"ddd-core-banking/internal/core-banking/application/usecases"
)

type transferProcessedPayload struct {
	TransferID        string `json:"transfer_id"`
	SenderAccountID   string `json:"sender_account_id"`
	ReceiverAccountID string `json:"receiver_account_id"`
	Amount            int64  `json:"amount"`
}

type TransferProcessedHandler struct {
	useCase *usecases.TransferBalanceUseCase
}

func NewTransferProcessedHandler(uc *usecases.TransferBalanceUseCase) *TransferProcessedHandler {
	return &TransferProcessedHandler{useCase: uc}
}

func (h *TransferProcessedHandler) Handle(payload []byte) error {
	var p transferProcessedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshaling TransferProcessed payload: %w", err)
	}
	return h.useCase.Execute(usecases.TransferBalanceInput{
		SenderAccountID:   p.SenderAccountID,
		ReceiverAccountID: p.ReceiverAccountID,
		Amount:            p.Amount,
	})
}
