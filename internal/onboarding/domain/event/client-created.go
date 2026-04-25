package event

import "time"

type ClientCreated struct {
	ClientID  string    `json:"client_id"`
	Document  string    `json:"document"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (e ClientCreated) EventName() string     { return "Onboarding.ClientCreated" }
func (e ClientCreated) OccurredAt() time.Time { return e.CreatedAt }
