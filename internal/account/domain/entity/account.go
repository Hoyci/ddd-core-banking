package entity

import (
	"time"

	"ddd-core-banking/internal/account/domain/errors"
	"ddd-core-banking/pkg/valueobjects"
)

type Account struct {
	clientID  string
	accountID string
	number    string
	blocked   *time.Time
	createdAt time.Time
}

type CreateAccountInput struct {
	ClientID string
	Number   string
}

func CreateAccount(input CreateAccountInput) (*Account, error) {
	if input.ClientID == "" {
		return nil, errors.ErrClientIDRequired
	}
	if input.Number == "" {
		return nil, errors.ErrAccountNumberRequired
	}

	return &Account{
		clientID:  input.ClientID,
		accountID: valueobjects.GenerateID(),
		number:    input.Number,
		blocked:   nil,
		createdAt: time.Now(),
	}, nil
}

type AccountData struct {
	AccountID string
	ClientID  string
	Number    string
	Blocked   *time.Time
	CreatedAt time.Time
}

func ReconstituteAccount(data AccountData) *Account {
	return &Account{
		accountID: data.AccountID,
		clientID:  data.ClientID,
		number:    data.Number,
		blocked:   data.Blocked,
		createdAt: data.CreatedAt,
	}
}

func (a *Account) AccountID() string    { return a.accountID }
func (a *Account) ClientID() string     { return a.clientID }
func (a *Account) Number() string       { return a.number }
func (a *Account) Blocked() *time.Time  { return a.blocked }
func (a *Account) CreatedAt() time.Time { return a.createdAt }
