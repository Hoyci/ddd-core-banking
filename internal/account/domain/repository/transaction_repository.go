package repository

import "ddd-core-banking/internal/account/domain/entity"

type TransactionRepository interface {
	Save(transaction *entity.Transaction) error
}
