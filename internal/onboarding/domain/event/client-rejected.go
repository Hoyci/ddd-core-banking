package event

import "time"

type ClientRejected struct {
	ClientID  string
	Document  string
	FullName  string
	Reason    string
	Email     string
	CreatedAt time.Time
}

func (e ClientRejected) EventName() string     { return "Onboarding.ClientRejected" }
func (e ClientRejected) OccurredAt() time.Time { return e.CreatedAt }
