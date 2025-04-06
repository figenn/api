package utils

func ValidateCurrency(currency string) bool {
	supportedCurrencies := []string{"USD", "EUR", "GBP", "CHF"}

	for _, c := range supportedCurrencies {
		if currency == c {
			return true
		}
	}
	return false
}
