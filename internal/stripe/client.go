package stripe

import (
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/client"
)

type stripeClient struct {
	client *client.API
}

func NewStripeClient(apiKey string) StripeService {
	sc := &client.API{}
	sc.Init(apiKey, nil)

	return &stripeClient{
		client: sc,
	}
}

func (s *stripeClient) CreateCheckoutSession(params *CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	sessionParams := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
			"revolut_pay",
			"paypal",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(params.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(params.SuccessURL),
		CancelURL:  stripe.String(params.CancelURL),
	}

	if params.CustomerEmail != "" {
		sessionParams.CustomerEmail = stripe.String(params.CustomerEmail)
	}

	return s.client.CheckoutSessions.New(sessionParams)
}

func (s *stripeClient) GetSubscription(subscriptionID string) (*stripe.Subscription, error) {
	return s.client.Subscriptions.Get(subscriptionID, nil)
}

func (s *stripeClient) CancelSubscription(subscriptionID string) (*stripe.Subscription, error) {
	return s.client.Subscriptions.Cancel(subscriptionID, nil)
}
