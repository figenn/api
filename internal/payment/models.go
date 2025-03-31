package payment

type CheckoutSessionParams struct {
	Plan       string `json:"price_id" form:"price_id"`
	CustomerId string `json:"customer_id" form:"customer_id"`
}
