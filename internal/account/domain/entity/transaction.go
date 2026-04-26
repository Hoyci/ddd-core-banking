package entity

import (
	"ddd-core-banking/internal/account/domain/errors"
	"ddd-core-banking/pkg/valueobjects"
	"time"
)

type TransactionType string

const (
	CREDIT TransactionType = "CREDIT"
	DEBIT  TransactionType = "DEBIT"
)

type Entry struct {
	accountID       string
	transactionType TransactionType
	amount          int64
}

func (e Entry) AccountID() string       { return e.accountID }
func (e Entry) TransactionType() string { return string(e.transactionType) }
func (e Entry) Amount() int64           { return e.amount }

type Transaction struct {
	transactionID string
	amount        int64
	createdAt     time.Time
	entries       [2]Entry
}

type CreateTransactionInput struct {
	SenderAccountID   string
	ReceiverAccountID string
	Amount            int64
}

func CreateTransaction(input CreateTransactionInput) (*Transaction, error) {
	if input.SenderAccountID == "" {
		return nil, errors.ErrSenderAccountIDRequired
	}
	if input.ReceiverAccountID == "" {
		return nil, errors.ErrReceiverAccountIDRequired
	}
	if input.Amount <= 0 {
		return nil, errors.ErrTransactionAmountInvalid
	}

	return &Transaction{
		transactionID: valueobjects.GenerateID(),
		amount:        input.Amount,
		createdAt:     time.Now(),
		entries: [2]Entry{
			{accountID: input.SenderAccountID, transactionType: DEBIT, amount: input.Amount},
			{accountID: input.ReceiverAccountID, transactionType: CREDIT, amount: input.Amount},
		},
	}, nil
}

func (t *Transaction) TransactionID() string { return t.transactionID }
func (t *Transaction) Amount() int64         { return t.amount }
func (t *Transaction) CreatedAt() time.Time  { return t.createdAt }
func (t *Transaction) Entries() [2]Entry     { return t.entries }
