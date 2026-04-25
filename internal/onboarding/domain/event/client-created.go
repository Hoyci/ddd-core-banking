package event

import "time"

type ClientCreated struct {
	ClientID  string
	Document  string
	FullName  string
	Email     string
	CreatedAt time.Time
}

func (e ClientCreated) EventName() string     { return "Onboarding.ClientCreated" }
func (e ClientCreated) OccurredAt() time.Time { return e.CreatedAt }
