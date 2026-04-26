package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"ddd-core-banking/internal/payment/application/usecases"
	payerrors "ddd-core-banking/internal/payment/domain/errors"
	pkghttp "ddd-core-banking/pkg/http"
)

type PaymentHandler struct {
	payInvoice    *usecases.PayInvoiceUseCase
	transferFunds *usecases.TransferFundsUseCase
}

func NewPaymentHandler(payInvoice *usecases.PayInvoiceUseCase, transferFunds *usecases.TransferFundsUseCase) *PaymentHandler {
	return &PaymentHandler{payInvoice: payInvoice, transferFunds: transferFunds}
}

func (h *PaymentHandler) PayInvoice(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccountID string `json:"account_id"`
		Barcode   string `json:"barcode"`
		Amount    int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, pkghttp.ApiResponse{Message: "invalid request body"})
		return
	}

	fmt.Println(body)

	output, err := h.payInvoice.Execute(context.Background(), usecases.PayInvoiceInput{
		AccountID: body.AccountID,
		Barcode:   body.Barcode,
		Amount:    body.Amount,
	})
	if err != nil {
		writeJSON(w, paymentStatusCode(err), pkghttp.ApiResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, pkghttp.ApiResponse{Data: output})
}

func (h *PaymentHandler) TransferFunds(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SenderAccountID   string `json:"sender_account_id"`
		ReceiverAccountID string `json:"receiver_account_id"`
		Amount            int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, pkghttp.ApiResponse{Message: "invalid request body"})
		return
	}

	output, err := h.transferFunds.Execute(context.Background(), usecases.TransferFundsInput{
		SenderAccountID:   body.SenderAccountID,
		ReceiverAccountID: body.ReceiverAccountID,
		Amount:            body.Amount,
	})
	if err != nil {
		writeJSON(w, paymentStatusCode(err), pkghttp.ApiResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, pkghttp.ApiResponse{Data: output})
}

func paymentStatusCode(err error) int {
	switch {
	case errors.Is(err, payerrors.ErrInsufficientFunds),
		errors.Is(err, payerrors.ErrAmountInvalid),
		errors.Is(err, payerrors.ErrSelfTransferNotAllowed),
		errors.Is(err, payerrors.ErrAccountIDRequired),
		errors.Is(err, payerrors.ErrBarcodeRequired),
		errors.Is(err, payerrors.ErrSenderAccountRequired),
		errors.Is(err, payerrors.ErrReceiverAccountRequired):
		return http.StatusUnprocessableEntity
	case errors.Is(err, payerrors.ErrAccountBlocked):
		return http.StatusForbidden
	case errors.Is(err, payerrors.ErrAccountNotFound):
		return http.StatusNotFound
	case errors.Is(err, payerrors.ErrCoreBankingUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, body pkghttp.ApiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
