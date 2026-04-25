package valueobjects

import (
	"ddd-core-banking/internal/onboarding/domain"
	"regexp"
	"strings"
)

var onlyAlpha = regexp.MustCompile(`[^a-zA-Z]`)

type AddressInput struct {
	ZipCode      string
	Street       string
	Number       string
	Complement   string
	Neighborhood string
	City         string
	State        string
}

type Address struct {
	zipCode      string
	street       string
	number       string
	complement   string
	neighborhood string
	city         string
	state        string
}

func NewAddress(input AddressInput) (Address, error) {
	zip := onlyDigits.ReplaceAllString(input.ZipCode, "")
	if len(zip) != 8 {
		return Address{}, domain.ErrInvalidZipCode
	}

	street := strings.TrimSpace(input.Street)
	if street == "" {
		return Address{}, domain.ErrStreetRequired
	}

	number := strings.TrimSpace(input.Number)
	if number == "" {
		return Address{}, domain.ErrAddressNumberRequired
	}

	neighborhood := strings.TrimSpace(input.Neighborhood)
	if neighborhood == "" {
		return Address{}, domain.ErrNeighborhoodRequired
	}

	city := strings.TrimSpace(input.City)
	if city == "" {
		return Address{}, domain.ErrCityRequired
	}

	state := strings.ToUpper(strings.TrimSpace(input.State))
	if state == "" {
		return Address{}, domain.ErrStateRequired
	}
	if len(onlyAlpha.ReplaceAllString(state, "")) != 2 || len(state) != 2 {
		return Address{}, domain.ErrInvalidState
	}

	return Address{
		zipCode:      zip,
		street:       street,
		number:       number,
		complement:   strings.TrimSpace(input.Complement),
		neighborhood: neighborhood,
		city:         city,
		state:        state,
	}, nil
}

func (a Address) ZipCode() string      { return a.zipCode }
func (a Address) Street() string       { return a.street }
func (a Address) Number() string       { return a.number }
func (a Address) Complement() string   { return a.complement }
func (a Address) Neighborhood() string { return a.neighborhood }
func (a Address) City() string         { return a.city }
func (a Address) State() string        { return a.state }
