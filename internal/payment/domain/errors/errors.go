package errors

import "errors"

var (
	ErrAccountIDRequired       = errors.New("account ID is required")
	ErrBarcodeRequired         = errors.New("barcode is required")
	ErrAmountInvalid           = errors.New("amount must be greater than zero")
	ErrSenderAccountRequired   = errors.New("sender account ID is required")
	ErrReceiverAccountRequired = errors.New("receiver account ID is required")
	ErrSelfTransferNotAllowed  = errors.New("self transfer is not allowed")
	ErrInsufficientFunds       = errors.New("insufficient funds")
	ErrAccountBlocked          = errors.New("account is blocked")
	ErrAccountNotFound         = errors.New("account not found")
	ErrCoreBankingUnavailable  = errors.New("core banking service temporarily unavailable")
)
