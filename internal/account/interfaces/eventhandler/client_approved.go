package eventhandler

import (
	"encoding/json"
	"fmt"

	"ddd-core-banking/internal/account/application/usecases"
)

type clientApprovedPayload struct {
	ClientID string `json:"client_id"`
}

type ClientApprovedHandler struct {
	useCase *usecases.CreateAccountUseCase
}

func NewClientApprovedHandler(uc *usecases.CreateAccountUseCase) *ClientApprovedHandler {
	return &ClientApprovedHandler{useCase: uc}
}

func (h *ClientApprovedHandler) Handle(payload []byte) error {
	var p clientApprovedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshaling ClientApproved payload: %w", err)
	}

	return h.useCase.Execute(usecases.CreateAccountInput{ClientID: p.ClientID})
}
