package valueobjects

import (
	"ddd-core-banking/internal/onboarding/domain"
	"regexp"
	"strconv"
)

type DocumentCategory string

const (
	CPF  DocumentCategory = "CPF"
	CNPJ DocumentCategory = "CNPJ"
)

type Document struct {
	number   string
	category DocumentCategory
}

var onlyDigits = regexp.MustCompile(`\D`)

func NewDocument(value string) (Document, error) {
	digits := onlyDigits.ReplaceAllString(value, "")

	switch len(digits) {
	case 11:
		if err := validateCPF(digits); err != nil {
			return Document{}, err
		}
		return Document{number: string(digits), category: CPF}, nil
	case 14:
		if err := validateCNPJ(digits); err != nil {
			return Document{}, err
		}
		return Document{number: string(digits), category: CNPJ}, nil
	default:
		return Document{}, domain.ErrInvalidDocument
	}
}

func validateCPF(digits string) error {
	if equalDigits(digits) {
		return domain.ErrInvalidCPF
	}

	if !validateCPFDigit(digits, 9) || !validateCPFDigit(digits, 10) {
		return domain.ErrInvalidCPF
	}

	return nil
}

func validateCNPJ(digits string) error {
	if equalDigits(digits) {
		return domain.ErrInvalidCNPJ
	}

	if !validateCNPJDigit(digits, 12) || !validateCPFDigit(digits, 13) {
		return domain.ErrInvalidCNPJ
	}

	return nil
}

func validateCPFDigit(digits string, pos int) bool {
	sum := 0
	weight := pos + 1

	for _, r := range digits[:pos] {
		digit := int(r - '0')
		sum += digit * weight
		weight--
	}

	remainder := sum % 11
	expected := 0

	if remainder >= 2 {
		expected = 11 - remainder
	}

	currentDigit, _ := strconv.Atoi(string(digits[pos]))
	return currentDigit == expected
}

func validateCNPJDigit(digits string, pos int) bool {
	weights := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	offset := 13 - pos
	sum := 0

	for i, r := range digits[:pos] {
		digit := int(r - '0')
		sum += digit * weights[offset+i]
	}

	remainder := sum % 11
	expected := 0

	if remainder >= 2 {
		expected = 11 - remainder
	}

	currentDigit, _ := strconv.Atoi(string(digits[pos]))
	return currentDigit == expected
}

func (d Document) Number() string             { return d.number }
func (d Document) Category() DocumentCategory { return d.category }

func equalDigits(s string) bool {
	for _, c := range s[1:] {
		if c != rune(s[0]) {
			return false
		}
	}
	return true
}
