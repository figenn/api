package stripe

type CheckoutSessionParams struct {
	PriceID    string `json:"price_id" form:"price_id"`
	CustomerId string `json:"customer_id" form:"customer_id"`
	SuccessURL string `json:"success_url" form:"success_url"`
	CancelURL  string `json:"cancel_url" form:"cancel_url"`
}
