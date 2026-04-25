package handler

import (
	"ddd-core-banking/internal/onboarding/application/usecases"
	"ddd-core-banking/internal/onboarding/domain"
	"ddd-core-banking/internal/onboarding/domain/entity"
	pkghttp "ddd-core-banking/pkg/http"
	"ddd-core-banking/pkg/valueobjects"
	"encoding/json"
	"errors"
	"net/http"
)

type ClientHandler struct {
	createClient  *usecases.CreateClientUseCase
	approveClient *usecases.ApproveClientUseCase
	rejectClient  *usecases.RejectClientUseCase
}

func NewClientHandler(
	createClient *usecases.CreateClientUseCase,
	approveClient *usecases.ApproveClientUseCase,
	rejectClient *usecases.RejectClientUseCase,
) *ClientHandler {
	return &ClientHandler{
		createClient:  createClient,
		approveClient: approveClient,
		rejectClient:  rejectClient,
	}
}

type createClientAddressRequest struct {
	ZipCode      string `json:"zip_code"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
}

type createClientRequest struct {
	FullName string                     `json:"full_name"`
	Document string                     `json:"document"`
	Email    string                     `json:"email"`
	Phone    string                     `json:"phone"`
	Address  createClientAddressRequest `json:"address"`
}

func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	req := &createClientRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(pkghttp.ApiResponse{Message: "invalid request body"})
		return
	}

	input := entity.CreateClientInput{
		FullName: req.FullName,
		Email:    req.Email,
		Document: req.Document,
		Phone:    req.Phone,
		Address: valueobjects.AddressInput{
			ZipCode:      req.Address.ZipCode,
			Street:       req.Address.Street,
			Number:       req.Address.Number,
			Complement:   req.Address.Complement,
			Neighborhood: req.Address.Neighborhood,
			City:         req.Address.City,
			State:        req.Address.State,
		},
	}

	if err := h.createClient.Execute(input); err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyInUse):
			w.WriteHeader(http.StatusConflict) // 409
		case errors.Is(err, domain.ErrInvalidDocument),
			errors.Is(err, domain.ErrInvalidEmail),
			errors.Is(err, domain.ErrFullNameRequired),
			errors.Is(err, domain.ErrPhoneRequired),
			errors.Is(err, domain.ErrInvalidZipCode),
			errors.Is(err, domain.ErrStreetRequired),
			errors.Is(err, domain.ErrAddressNumberRequired),
			errors.Is(err, domain.ErrNeighborhoodRequired),
			errors.Is(err, domain.ErrCityRequired),
			errors.Is(err, domain.ErrStateRequired),
			errors.Is(err, domain.ErrInvalidState):
			w.WriteHeader(http.StatusUnprocessableEntity) // 422
		default:
			w.WriteHeader(http.StatusInternalServerError) // 500
		}
		json.NewEncoder(w).Encode(pkghttp.ApiResponse{Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pkghttp.ApiResponse{Message: "client successfully created"})
}

func (h *ClientHandler) Approve(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientID")
	if clientID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(pkghttp.ApiResponse{Message: "missing clientID"})
		return
	}

	input := usecases.ApproveClientInput{
		ClientID: clientID,
	}

	if err := h.approveClient.Execute(input); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, domain.ErrClientNotPending):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(pkghttp.ApiResponse{Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pkghttp.ApiResponse{Message: "client approved"})
}

type rejectClientRequest struct {
	Reason string `json:"reason"`
}

func (h *ClientHandler) Reject(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientID")
	if clientID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(pkghttp.ApiResponse{Message: "missing clientID"})
		return
	}

	req := &rejectClientRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(pkghttp.ApiResponse{Message: "invalid request body"})
		return
	}

	input := usecases.RejectClientInput{
		ClientID: clientID,
		Reason:   req.Reason,
	}

	if err := h.rejectClient.Execute(input); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, domain.ErrClientNotPending),
			errors.Is(err, domain.ErrRejectionReasonRequired):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(pkghttp.ApiResponse{Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pkghttp.ApiResponse{Message: "client rejected"})
}
