package valueobjects

import "fmt"

func NewAccountNumber(seq int64) string {
	return generateAccountNumber(seq)
}

func generateAccountNumber(seq int64) string {
	base := fmt.Sprintf("%09d", seq)

	dv := calculateDV(base)

	return fmt.Sprintf("%s-%d", base, dv)
}

func calculateDV(number string) int {
	sum := 0
	multiplier := 2

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		result := digit * multiplier

		if result > 9 {
			result = (result / 10) + (result % 10)
		}

		sum += result

		if multiplier == 2 {
			multiplier = 1
		} else {
			multiplier = 2
		}
	}

	return (10 - (sum % 10)) % 10
}
