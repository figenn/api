package utils

import (
	"fmt"
	"strconv"
)

func ValidateYear(year string) (int, error) {
	y, err := strconv.Atoi(year)
	if err != nil || y < 2020 || y > 2100 {
		return 0, fmt.Errorf("Invalid year. Must be between 2020 and 2100")
	}
	return y, nil
}

func ValidateMonth(month string) (int, error) {
	m, err := strconv.Atoi(month)
	if err != nil || m < 1 || m > 12 {
		return 0, fmt.Errorf("Invalid month. Must be between 1 and 12")
	}
	return m, nil
}
