package repository

import "ddd-core-banking/internal/account/domain/entity"

type AccountRepository interface {
	Save(account *entity.Account) error
	FindByID(id string) (*entity.Account, error)
	FindByClientID(clientID string) (*entity.Account, error)
	GetBalanceByAccountID(accountID string) (int64, error)
}
