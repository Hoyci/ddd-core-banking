package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ddd-core-banking/internal/onboarding/domain"
	"ddd-core-banking/internal/onboarding/domain/entity"
	"ddd-core-banking/pkg/events"
	"ddd-core-banking/pkg/valueobjects"
)

type ClientRepository struct {
	db *pgxpool.Pool
}

func NewClientRepository(db *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) Save(client *entity.Client, domainEvents []events.DomainEvent) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		INSERT INTO clients (id, full_name, email, phone, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status
	`,
		client.ID(),
		client.FullName(),
		client.Email(),
		client.Phone(),
		string(client.Status()),
		client.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("upserting client: %w", err)
	}

	_, err = tx.Exec(context.Background(), `
		INSERT INTO client_documents (client_id, number, type)
		VALUES ($1, $2, $3)
		ON CONFLICT (client_id) DO NOTHING
	`,
		client.ID(),
		client.Document().Number(),
		string(client.Document().Category()),
	)
	if err != nil {
		return fmt.Errorf("upserting client document: %w", err)
	}

	addr := client.Address()
	var complement *string
	if c := addr.Complement(); c != "" {
		complement = &c
	}

	_, err = tx.Exec(context.Background(), `
		INSERT INTO client_addresses (client_id, zip_code, street, number, complement, neighborhood, city, state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (client_id) DO UPDATE SET
			zip_code     = EXCLUDED.zip_code,
			street       = EXCLUDED.street,
			number       = EXCLUDED.number,
			complement   = EXCLUDED.complement,
			neighborhood = EXCLUDED.neighborhood,
			city         = EXCLUDED.city,
			state        = EXCLUDED.state
	`,
		client.ID(),
		addr.ZipCode(),
		addr.Street(),
		addr.Number(),
		complement,
		addr.Neighborhood(),
		addr.City(),
		addr.State(),
	)
	if err != nil {
		return fmt.Errorf("upserting client address: %w", err)
	}

	for _, evt := range domainEvents {
		payload, err := json.Marshal(evt)
		if err != nil {
			return fmt.Errorf("marshaling event %s: %w", evt.EventName(), err)
		}

		_, err = tx.Exec(context.Background(), `
			INSERT INTO outbox (event_name, payload, occurred_at)
			VALUES ($1, $2, $3)
		`,
			evt.EventName(),
			payload,
			evt.OccurredAt(),
		)
		if err != nil {
			return fmt.Errorf("inserting outbox record for %s: %w", evt.EventName(), err)
		}
	}

	_, err = tx.Exec(context.Background(), `SELECT pg_notify('outbox_event', '')`)
	if err != nil {
		return fmt.Errorf("notifying outbox channel: %w", err)
	}

	return tx.Commit(context.Background())
}

func (r *ClientRepository) FindByID(id string) (*entity.Client, error) {
	row := r.db.QueryRow(context.Background(), `
		SELECT
			c.id, c.full_name, c.email, c.phone, c.status, c.created_at,
			d.number AS document_number, d.type AS document_type,
			a.zip_code, a.street, a.number AS address_number, a.complement,
			a.neighborhood, a.city, a.state
		FROM clients c
		JOIN client_documents d ON d.client_id = c.id
		JOIN client_addresses  a ON a.client_id = c.id
		WHERE c.id = $1
	`, id)

	return scanClient(row)
}

func (r *ClientRepository) FindByEmail(email string) (*entity.Client, error) {
	row := r.db.QueryRow(context.Background(), `
		SELECT
			c.id, c.full_name, c.email, c.phone, c.status, c.created_at,
			d.number AS document_number, d.type AS document_type,
			a.zip_code, a.street, a.number AS address_number, a.complement,
			a.neighborhood, a.city, a.state
		FROM clients c
		JOIN client_documents d ON d.client_id = c.id
		JOIN client_addresses  a ON a.client_id = c.id
		WHERE c.email = $1
	`, email)

	return scanClient(row)
}

func scanClient(row pgx.Row) (*entity.Client, error) {
	var (
		id             string
		fullName       string
		emailStr       string
		phone          string
		statusStr      string
		createdAt      time.Time
		documentNumber string
		documentType   string
		zipCode        string
		street         string
		addressNumber  string
		complement     *string
		neighborhood   string
		city           string
		state          string
	)

	err := row.Scan(
		&id, &fullName, &emailStr, &phone, &statusStr, &createdAt,
		&documentNumber, &documentType,
		&zipCode, &street, &addressNumber, &complement,
		&neighborhood, &city, &state,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanning client row: %w", err)
	}

	doc, err := valueobjects.NewDocument(documentNumber)
	if err != nil {
		return nil, fmt.Errorf("reconstituting document: %w", err)
	}

	email, err := valueobjects.NewEmail(emailStr)
	if err != nil {
		return nil, fmt.Errorf("reconstituting email: %w", err)
	}

	complementStr := ""
	if complement != nil {
		complementStr = *complement
	}

	address, err := valueobjects.NewAddress(valueobjects.AddressInput{
		ZipCode:      zipCode,
		Street:       street,
		Number:       addressNumber,
		Complement:   complementStr,
		Neighborhood: neighborhood,
		City:         city,
		State:        state,
	})
	if err != nil {
		return nil, fmt.Errorf("reconstituting address: %w", err)
	}

	return entity.ReconstituteClient(entity.ClientData{
		ID:        id,
		Document:  doc,
		FullName:  fullName,
		Email:     email,
		Phone:     phone,
		Address:   address,
		Status:    entity.OnboardingStatus(statusStr),
		CreatedAt: createdAt,
	}), nil
}
