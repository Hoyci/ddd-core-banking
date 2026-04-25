package domain

import "errors"

var (
	ErrNotFound                = errors.New("not found")
	ErrEmailAlreadyInUse       = errors.New("email already in use")
	ErrInvalidDocument         = errors.New("invalid document")
	ErrFullNameRequired        = errors.New("full name is required")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrEmailIsRequired         = errors.New("email is required")
	ErrExceedLength            = errors.New("exceeds maximum allowed length")
	ErrPhoneRequired           = errors.New("phone is required")
	ErrClientNotPending        = errors.New("client is not in pending status")
	ErrRejectionReasonRequired = errors.New("rejection reason is required")
	ErrInvalidCPF              = errors.New("invalid CPF")
	ErrInvalidCNPJ             = errors.New("invalid CNPJ")
	ErrInvalidZipCode          = errors.New("invalid zip code")
	ErrStreetRequired          = errors.New("street is required")
	ErrAddressNumberRequired   = errors.New("address number is required")
	ErrNeighborhoodRequired    = errors.New("neighborhood is required")
	ErrCityRequired            = errors.New("city is required")
	ErrStateRequired           = errors.New("state is required")
	ErrInvalidState            = errors.New("state must be a 2-letter UF code")
)
