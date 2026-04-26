package event

import "time"

type TransferProcessed struct {
	TransferID        string    `json:"transfer_id"`
	SenderAccountID   string    `json:"sender_account_id"`
	ReceiverAccountID string    `json:"receiver_account_id"`
	Amount            int64     `json:"amount"`
	CreatedAt         time.Time `json:"created_at"`
}

func NewTransferProcessed(transferID, senderID, receiverID string, amount int64) TransferProcessed {
	return TransferProcessed{
		TransferID:        transferID,
		SenderAccountID:   senderID,
		ReceiverAccountID: receiverID,
		Amount:            amount,
		CreatedAt:         time.Now(),
	}
}

func (e TransferProcessed) EventName() string     { return "Payment.TransferProcessed" }
func (e TransferProcessed) OccurredAt() time.Time { return e.CreatedAt }
