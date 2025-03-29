package subscriptions

import "errors"

var (
	ErrUserIDAndSubIDRequired = errors.New("user ID and subscription ID are required")
	ErrSubscriptionNotFound   = errors.New("subscription not found")
	ErrUserPermissionDenied   = errors.New("user not authorized to access this subscription")
	ErrNoFieldsToUpdate       = errors.New("no fields provided for update")
	ErrInvalidRequestFormat   = errors.New("invalid request format")
	ErrInvalidUserSession     = errors.New("invalid user session")
	ErrInvalidSubscriptionID  = errors.New("invalid subscription ID")
	ErrFailedCreateSub        = errors.New("failed to create subscription")
	ErrInvalidPeriod          = errors.New("invalid period")
	ErrInvalidWeek            = errors.New("invalid week")
)
