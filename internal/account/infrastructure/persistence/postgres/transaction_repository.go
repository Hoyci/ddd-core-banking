package postgres

import (
	"context"
	"ddd-core-banking/internal/account/domain/entity"
	"ddd-core-banking/pkg/valueobjects"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Save(transaction *entity.Transaction) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		INSERT INTO transactions (id, amount, created_at)
		VALUES ($1, $2, $3)
	`,
		transaction.TransactionID(),
		transaction.Amount(),
		transaction.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("inserting transaction: %w", err)
	}

	entries := transaction.Entries()

	for _, entry := range entries {
		_, err = tx.Exec(context.Background(), `
			INSERT INTO entries (id, transaction_id, account_id, type, amount, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
			valueobjects.GenerateID(),
			transaction.TransactionID(),
			entry.AccountID(),
			entry.TransactionType(),
			entry.Amount(),
			transaction.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("inserting entry: %w", err)
		}
	}

	_, err = tx.Exec(context.Background(), `
		UPDATE accounts SET balance = balance - $1 WHERE id = $2
	`,
		transaction.Amount(),
		entries[0].AccountID(),
	)
	if err != nil {
		return fmt.Errorf("updating sender balance: %w", err)
	}

	_, err = tx.Exec(context.Background(), `
		UPDATE accounts SET balance = balance + $1 WHERE id = $2
	`,
		transaction.Amount(),
		entries[1].AccountID(),
	)
	if err != nil {
		return fmt.Errorf("updating receiver balance: %w", err)
	}

	return tx.Commit(context.Background())
}
