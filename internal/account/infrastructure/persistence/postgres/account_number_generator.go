package postgres

import (
	"context"
	"fmt"

	"ddd-core-banking/pkg/valueobjects"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SequenceAccountNumberGenerator struct {
	db *pgxpool.Pool
}

func NewSequenceAccountNumberGenerator(db *pgxpool.Pool) *SequenceAccountNumberGenerator {
	return &SequenceAccountNumberGenerator{db: db}
}

func (g *SequenceAccountNumberGenerator) Next() (string, error) {
	var seq int64
	err := g.db.QueryRow(context.Background(), "SELECT nextval('account_number_seq')").Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("fetching next account number sequence: %w", err)
	}
	return valueobjects.NewAccountNumber(seq), nil
}
