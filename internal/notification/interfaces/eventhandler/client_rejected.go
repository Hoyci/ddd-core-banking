package eventhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ddd-core-banking/internal/notification/application/usecases"
)

type clientRejectedPayload struct {
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type ClientRejectedHandler struct {
	useCase *usecases.NotifyRejectionUseCase
}

func NewClientRejectedHandler(uc *usecases.NotifyRejectionUseCase) *ClientRejectedHandler {
	return &ClientRejectedHandler{useCase: uc}
}

func (h *ClientRejectedHandler) Handle(payload []byte) error {
	var p clientRejectedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshaling ClientRejected payload: %w", err)
	}
	return h.useCase.Execute(context.Background(), usecases.NotifyRejectionInput{
		FullName: p.FullName,
		Email:    p.Email,
		Reason:   p.Reason,
	})
}
