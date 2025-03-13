package stripe

type CheckoutSessionParams struct {
	PriceID       string `json:"price_id" form:"price_id"`
	CustomerEmail string `json:"customer_email" form:"customer_email"`
	SuccessURL    string `json:"success_url" form:"success_url"`
	CancelURL     string `json:"cancel_url" form:"cancel_url"`
}
