package payment

import "errors"

var (
	ErrMissingPriceID = errors.New("missing price ID")
	ErrInvalidPriceID = errors.New("invalid price ID")
)
