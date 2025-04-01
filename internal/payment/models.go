package payment

import (
	"time"

	"github.com/gofrs/uuid"
)

type SubscriptionType string

const (
	Free         SubscriptionType = "free"
	Premium      SubscriptionType = "premium"
	Professional SubscriptionType = "professional"
)

type UserSubscription struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	StripeSubscriptionID string
	StripePriceID        string
	SubscriptionType     SubscriptionType
	Status               string
	CancelAtPeriodEnd    bool
	CurrentPeriodStart   time.Time
	CurrentPeriodEnd     time.Time
	CanceledAt           *time.Time
	EndsAt               *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type CheckoutSessionParams struct {
	Plan       string `json:"plan" form:"plan"`
	CustomerId string `json:"customer_id" form:"customer_id"`
}

const (
	PremiumPriceID = "price_1R1zTpG72A5CyjpR5Iw2sQJH"
	ProPriceID     = "price_1R26rXG72A5CyjpRjnylykuT"
)

var planMap = map[string]SubscriptionType{
	PremiumPriceID: Premium,
	ProPriceID:     Professional,
}
