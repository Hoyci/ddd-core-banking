package errors

import "errors"

var (
	ErrClientIDRequired          = errors.New("clientID is required")
	ErrAccountNumberRequired     = errors.New("account number is required")
	ErrAccountNotFound           = errors.New("account not found")
	ErrSenderAccountIDRequired   = errors.New("sender's accountID is required")
	ErrReceiverAccountIDRequired = errors.New("receiver's accountID is required")
	ErrTransactionAmountInvalid  = errors.New("transaction amount is less or equal to 0")
	ErrAccountBlocked            = errors.New("account is blocked")
	ErrSelfTransferNotAllowed    = errors.New("self transfer is not allowed")
	ErrInsufficientFunds         = errors.New("insufficient funds")
)
