package event

import "time"

type ClientApproved struct {
	ClientID  string    `json:"client_id"`
	Document  string    `json:"document"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (e ClientApproved) EventName() string     { return "Onboarding.ClientApproved" }
func (e ClientApproved) OccurredAt() time.Time { return e.CreatedAt }
