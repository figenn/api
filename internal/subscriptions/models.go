package subscriptions

import "time"

type Subscription struct {
	Id              string     `json:"id"`
	UserId          string     `json:"user_id"`
	Name            string     `json:"name"`
	Category        string     `json:"category"`
	Color           string     `json:"color"`
	Description     string     `json:"description"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	Price           float64    `json:"price"`
	LogoUrl         string     `json:"logo_url"`
	Active          bool       `json:"active"`
	NextBillingDate time.Time  `json:"next_billing_date"`
	IsRecuring      bool       `json:"is_recurring"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateSubscriptionRequest struct {
	Name            string     `json:"name" form:"name"`
	Category        string     `json:"category" form:"category"`
	Color           string     `json:"color" form:"color"`
	Description     string     `json:"description" form:"description"`
	StartDate       *time.Time `json:"start_date" form:"start_date"`
	EndDate         *time.Time `json:"end_date" form:"end_date"`
	Price           float64    `json:"price" form:"price"`
	NextBillingDate time.Time  `json:"next_billing_date" form:"next_billing_date"`
	IsRecuring      bool       `json:"is_recurring" form:"is_recurring"`
}

type LogoResponse struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Route    string `json:"route"`
	URL      string `json:"url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
