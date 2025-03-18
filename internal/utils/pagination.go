package utils

import (
	"errors"
	"strconv"

	"github.com/labstack/echo/v4"
)

func GetPaginationParams(c echo.Context) (int, int, error) {
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	if limitStr == "" {
		limitStr = "10"
	}

	if pageStr == "" {
		pageStr = "1"
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		return 0, 0, errors.New("Invalid limit")
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 0, 0, errors.New("Invalid page number")
	}

	offset := (page - 1) * limit

	return limit, offset, nil
}
