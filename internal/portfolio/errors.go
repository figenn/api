package portfolio

import "errors"

var (
	ErrNoMatchingStock  = errors.New("no matching stock found")
	ErrOverviewNotFound = errors.New("stock overview not available")
)
