package payment

type CheckoutSessionParams struct {
	Plan       string `json:"plan" form:"plan"`
	CustomerId string `json:"customer_id" form:"customer_id"`
}
