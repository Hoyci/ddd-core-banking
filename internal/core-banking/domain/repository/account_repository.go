package repository

import "ddd-core-banking/internal/core-banking/domain/entity"

type AccountRepository interface {
	Save(account *entity.Account) error
	FindByID(id string) (*entity.Account, error)
	FindByClientID(clientID string) (*entity.Account, error)
	Debit(accountID string, amount int64) error
	TransferBalance(senderAccountID, receiverAccountID string, amount int64) error
}
