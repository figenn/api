package payment

import "errors"

var (
	ErrMissingPriceID             = errors.New("missing price ID")
	ErrInvalidPriceID             = errors.New("invalid price ID")
	ErrNotFound                   = errors.New("not found")
	ErrInvalidInvoicePayload      = errors.New("invalid invoice payload")
	ErrNoSubscriptionLineItem     = errors.New("invoice has no subscription line item")
	ErrUnknownPriceID             = errors.New("unknown price ID")
	ErrInvalidSubscriptionPayload = errors.New("invalid subscription payload")
	ErrInvalidSubscriptionData    = errors.New("subscription has no valid price")
	ErrCheckoutSessionInvalid     = errors.New("missing customer or subscription")
	ErrStripeSubscriptionFetch    = errors.New("unable to fetch subscription from Stripe")
	ErrUpdateSubscriptionFailed   = errors.New("error updating subscription")
)
