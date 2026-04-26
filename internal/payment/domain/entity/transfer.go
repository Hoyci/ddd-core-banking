package entity

import (
	"time"

	payerrors "ddd-core-banking/internal/payment/domain/errors"
	"ddd-core-banking/pkg/valueobjects"
)

type EntryType string

const (
	Debit  EntryType = "DEBIT"
	Credit EntryType = "CREDIT"
)

type Entry struct {
	accountID string
	entryType EntryType
	amount    int64
}

func (e Entry) AccountID() string { return e.accountID }
func (e Entry) EntryType() string { return string(e.entryType) }
func (e Entry) Amount() int64     { return e.amount }

type Transfer struct {
	transferID string
	amount     int64
	createdAt  time.Time
	entries    [2]Entry
}

type CreateTransferInput struct {
	SenderAccountID   string
	ReceiverAccountID string
	Amount            int64
}

func CreateTransfer(input CreateTransferInput) (*Transfer, error) {
	if input.SenderAccountID == "" {
		return nil, payerrors.ErrSenderAccountRequired
	}
	if input.ReceiverAccountID == "" {
		return nil, payerrors.ErrReceiverAccountRequired
	}
	if input.SenderAccountID == input.ReceiverAccountID {
		return nil, payerrors.ErrSelfTransferNotAllowed
	}
	if input.Amount <= 0 {
		return nil, payerrors.ErrAmountInvalid
	}
	return &Transfer{
		transferID: valueobjects.GenerateID(),
		amount:     input.Amount,
		createdAt:  time.Now(),
		entries: [2]Entry{
			{accountID: input.SenderAccountID, entryType: Debit, amount: input.Amount},
			{accountID: input.ReceiverAccountID, entryType: Credit, amount: input.Amount},
		},
	}, nil
}

func (t *Transfer) TransferID() string   { return t.transferID }
func (t *Transfer) Amount() int64        { return t.amount }
func (t *Transfer) CreatedAt() time.Time { return t.createdAt }
func (t *Transfer) Entries() [2]Entry    { return t.entries }
