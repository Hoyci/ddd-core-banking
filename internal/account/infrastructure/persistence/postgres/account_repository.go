package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	accounterrors "ddd-core-banking/internal/account/domain/errors"
	"ddd-core-banking/internal/account/domain/entity"
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
		SELECT id, client_id, number, blocked, created_at
		FROM accounts WHERE id = $1
	`, id)
	return scanAccount(row)
}

func (r *AccountRepository) FindByClientID(clientID string) (*entity.Account, error) {
	row := r.db.QueryRow(context.Background(), `
		SELECT id, client_id, number, blocked, created_at
		FROM accounts WHERE client_id = $1
	`, clientID)
	return scanAccount(row)
}

func scanAccount(row pgx.Row) (*entity.Account, error) {
	var (
		id        string
		clientID  string
		number    string
		blocked   *time.Time
		createdAt time.Time
	)

	err := row.Scan(&id, &clientID, &number, &blocked, &createdAt)
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
		Blocked:   blocked,
		CreatedAt: createdAt,
	}), nil
}
