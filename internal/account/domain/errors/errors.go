package errors

import "errors"

var (
	ErrClientIDRequired      = errors.New("clientID is required")
	ErrAccountNumberRequired = errors.New("account number is required")
	ErrAccountNotFound       = errors.New("account not found")
)
