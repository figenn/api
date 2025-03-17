package utils

import (
	"errors"
	"strconv"

	"github.com/labstack/echo/v4"
)

func GetPaginationParams(c echo.Context) (int, int, error) {
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	if limitStr == "" {
		limitStr = strconv.Itoa(10)
	}
	if offsetStr == "" {
		offsetStr = "0"
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		return 0, 0, errors.New("Invalid limit")
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return 0, 0, errors.New("Invalid offset")
	}

	return limit, offset, nil
}
