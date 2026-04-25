package event

import "time"

type ClientRejected struct {
	ClientID  string    `json:"client_id"`
	Document  string    `json:"document"`
	FullName  string    `json:"full_name"`
	Reason    string    `json:"reason"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (e ClientRejected) EventName() string     { return "Onboarding.ClientRejected" }
func (e ClientRejected) OccurredAt() time.Time { return e.CreatedAt }
