package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ddd-core-banking/internal/core-banking/domain/entity"
	accounterrors "ddd-core-banking/internal/core-banking/domain/errors"
)

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Save(account *entity.Account) error {
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO accounts (id, client_id, number, blocked, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			blocked = EXCLUDED.blocked
	`,
		account.AccountID(),
		account.ClientID(),
		account.Number(),
		account.Blocked(),
		account.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("upserting account: %w", err)
	}
	return nil
}

func (r *AccountRepository) FindByID(id string) (*entity.Account, error) {
	row := r.db.QueryRow(context.Background(), `
		SELECT id, client_id, number, balance, blocked, created_at
		FROM accounts WHERE id = $1
	`, id)
	return scanAccount(row)
}

func (r *AccountRepository) FindByClientID(clientID string) (*entity.Account, error) {
	row := r.db.QueryRow(context.Background(), `
		SELECT id, client_id, number, balance, blocked, created_at
		FROM accounts WHERE client_id = $1
	`, clientID)
	return scanAccount(row)
}

func (r *AccountRepository) Debit(accountID string, amount int64) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE accounts SET balance = balance - $1 WHERE id = $2
	`, amount, accountID)
	if err != nil {
		return fmt.Errorf("debiting account %s: %w", accountID, err)
	}
	return nil
}

func (r *AccountRepository) TransferBalance(senderAccountID, receiverAccountID string, amount int64) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		UPDATE accounts SET balance = balance - $1 WHERE id = $2
	`, amount, senderAccountID)
	if err != nil {
		return fmt.Errorf("debiting sender %s: %w", senderAccountID, err)
	}

	_, err = tx.Exec(context.Background(), `
		UPDATE accounts SET balance = balance + $1 WHERE id = $2
	`, amount, receiverAccountID)
	if err != nil {
		return fmt.Errorf("crediting receiver %s: %w", receiverAccountID, err)
	}

	return tx.Commit(context.Background())
}

func scanAccount(row pgx.Row) (*entity.Account, error) {
	var (
		id        string
		clientID  string
		number    string
		balance   int64
		blocked   *time.Time
		createdAt time.Time
	)

	err := row.Scan(&id, &clientID, &number, &balance, &blocked, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, accounterrors.ErrAccountNotFound
		}
		return nil, fmt.Errorf("scanning account row: %w", err)
	}

	return entity.ReconstituteAccount(entity.AccountData{
		AccountID: id,
		ClientID:  clientID,
		Number:    number,
		Balance:   balance,
		Blocked:   blocked,
		CreatedAt: createdAt,
	}), nil
}
