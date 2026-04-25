package event

import "time"

type ClientApproved struct {
	ClientID  string
	Document  string
	FullName  string
	Email     string
	CreatedAt time.Time
}

func (e ClientApproved) EventName() string     { return "Onboarding.ClientApproved" }
func (e ClientApproved) OccurredAt() time.Time { return e.CreatedAt }
