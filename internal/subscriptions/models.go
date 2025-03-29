package subscriptions

import "time"

type Subscription struct {
	Id           string           `json:"id"`
	UserId       string           `json:"user_id"`
	Name         string           `json:"name"`
	Category     string           `json:"category"`
	Color        string           `json:"color"`
	Description  string           `json:"description"`
	StartDate    time.Time        `json:"start_date"`
	EndDate      *time.Time       `json:"end_date"`
	Price        float64          `json:"price"`
	LogoUrl      *string          `json:"logo_url"`
	IsActive     bool             `json:"is_active"`
	BillingCycle BillingCycleType `json:"billing_cycle"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type BillingCycleType string

const (
	Monthly   BillingCycleType = "monthly"
	Quarterly BillingCycleType = "quarterly"
	Annual    BillingCycleType = "annual"
)

type CreateSubscriptionRequest struct {
	Name         string           `json:"name" form:"name"`
	Category     string           `json:"category" form:"category"`
	Color        string           `json:"color" form:"color"`
	Description  string           `json:"description" form:"description"`
	StartDate    *time.Time       `json:"start_date" form:"start_date"`
	EndDate      *time.Time       `json:"end_date" form:"end_date"`
	Price        float64          `json:"price" form:"price"`
	LogoUrl      string           `json:"logo_url" form:"logo_url"`
	BillingCycle BillingCycleType `json:"billing_cycle" form:"billing_cycle"`
	IsRecuring   bool             `json:"is_recuring" form:"is_recuring"`
}

type LogoResponse struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Route    string `json:"route"`
	URL      string `json:"url"`
}

type UpdateSubscriptionRequest struct {
	Name        *string    `json:"name,omitempty" form:"name"`
	Category    *string    `json:"category,omitempty" form:"category"`
	Color       *string    `json:"color,omitempty" form:"color"`
	Description *string    `json:"description,omitempty" form:"description"`
	StartDate   *time.Time `json:"start_date,omitempty" form:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" form:"end_date"`
	Price       *float64   `json:"price,omitempty" form:"price"`
	IsActive    *bool      `json:"is_active,omitempty" form:"is_active"`
	IsRecuring  *bool      `json:"is_recuring,omitempty" form:"is_recuring"`
}

type SubscriptionCategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}
