package entity

import (
	"fmt"
	"time"

	"ddd-core-banking/internal/onboarding/domain"
	"ddd-core-banking/internal/onboarding/domain/event"
	"ddd-core-banking/pkg/events"
	"ddd-core-banking/pkg/valueobjects"
)

type OnboardingStatus string

const (
	OnboardingStatusPending  OnboardingStatus = "PENDING"
	OnboardingStatusApproved OnboardingStatus = "APPROVED"
	OnboardingStatusRejected OnboardingStatus = "REJECTED"
)

type Client struct {
	id        string
	document  valueobjects.Document
	fullName  string
	email     valueobjects.Email
	phone     string
	address   valueobjects.Address
	status    OnboardingStatus
	createdAt time.Time

	events []events.DomainEvent
}

type CreateClientInput struct {
	Document string
	FullName string
	Email    string
	Phone    string
	Address  valueobjects.AddressInput
}

type ClientData struct {
	ID        string
	Document  valueobjects.Document
	FullName  string
	Email     valueobjects.Email
	Phone     string
	Address   valueobjects.Address
	Status    OnboardingStatus
	CreatedAt time.Time
}

func CreateClient(input CreateClientInput) (*Client, error) {
	if input.FullName == "" {
		return nil, domain.ErrFullNameRequired
	}

	doc, err := valueobjects.NewDocument(input.Document)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidDocument, err)
	}

	email, err := valueobjects.NewEmail(input.Email)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidEmail, err)
	}

	if input.Phone == "" {
		return nil, domain.ErrPhoneRequired
	}

	address, err := valueobjects.NewAddress(input.Address)
	if err != nil {
		return nil, err
	}

	clientID := valueobjects.GenerateID()
	now := time.Now()

	client := &Client{
		id:        clientID,
		document:  doc,
		fullName:  input.FullName,
		email:     email,
		phone:     input.Phone,
		address:   address,
		status:    OnboardingStatusPending,
		createdAt: now,
		events: []events.DomainEvent{
			event.ClientCreated{
				ClientID:  clientID,
				Document:  doc.Number(),
				FullName:  input.FullName,
				Email:     email.Value(),
				CreatedAt: now,
			},
		},
	}

	return client, nil
}

func (c *Client) ApproveClient() error {
	if c.status != OnboardingStatusPending {
		return domain.ErrClientNotPending
	}

	c.status = OnboardingStatusApproved
	c.events = append(c.events, event.ClientApproved{
		ClientID:  c.id,
		Email:     c.Email(),
		Document:  c.Document().Number(),
		FullName:  c.FullName(),
		CreatedAt: time.Now(),
	})

	return nil
}

func (c *Client) RejectClient(reason string) error {
	if c.status != OnboardingStatusPending {
		return domain.ErrClientNotPending
	}

	if reason == "" {
		return domain.ErrRejectionReasonRequired
	}

	c.status = OnboardingStatusRejected
	c.events = append(c.events, event.ClientRejected{
		Email:     c.Email(),
		ClientID:  c.id,
		Document:  c.Document().Number(),
		FullName:  c.fullName,
		Reason:    reason,
		CreatedAt: time.Now(),
	})

	return nil
}

// Pensar numa melhoria para esse PullEvents para que não aconteça de apagar os eventos e o Save falhar
func (c *Client) PullEvents() []events.DomainEvent {
	evts := c.events
	c.events = nil
	return evts
}

func ReconstituteClient(data ClientData) *Client {
	return &Client{
		id:        data.ID,
		document:  data.Document,
		fullName:  data.FullName,
		email:     data.Email,
		phone:     data.Phone,
		address:   data.Address,
		status:    data.Status,
		createdAt: data.CreatedAt,
	}
}

func (c *Client) ID() string                      { return c.id }
func (c *Client) Document() valueobjects.Document { return c.document }
func (c *Client) FullName() string                { return c.fullName }
func (c *Client) Email() string                   { return c.email.Value() }
func (c *Client) Phone() string                   { return c.phone }
func (c *Client) Address() valueobjects.Address   { return c.address }
func (c *Client) Status() OnboardingStatus        { return c.status }
func (c *Client) CreatedAt() time.Time            { return c.createdAt }
