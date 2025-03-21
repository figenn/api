package subscriptions

import "errors"

var (
	ErrUserIDAndSubIDRequired = errors.New("Both user ID and subscription ID are required")
	ErrSubscriptionNotFound   = errors.New("Subscription not found for the given user ID and subscription ID")
	ErrUserPermissionDenied   = errors.New("User does not have permission to update this subscription")
	ErrNoFieldsToUpdate       = errors.New("No fields have been provided for update")
	ErrInvalidRequestFormat   = errors.New("The request format is invalid")
	ErrInvalidUserSession     = errors.New("Invalid user session, please log in again")
	ErrInvalidSubscriptionID  = errors.New("he provided subscription ID is invalid")
	ErrFailedCreateSub        = errors.New("Failed to create subscription")
	ErrInvalidPeriod          = errors.New("Invalid period provided")
)
