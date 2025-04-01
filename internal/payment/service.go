package payment

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/client"
	"github.com/stripe/stripe-go/v81/customer"
)

type PaymentService interface {
	CreateCheckoutSession(params *CheckoutSessionParams) (*stripe.CheckoutSession, error)
	GetSubscription(subscriptionID string) (*stripe.Subscription, error)
	CancelSubscription(subscriptionID string) (*stripe.Subscription, error)
	HandleWebhook(c echo.Context) error
}

type Service struct {
	client  *client.API
	r       *Repository
	appUrl  string
	planMap map[string]string
}

func NewService(apiKey string, repo *Repository) *Service {
	sc := &client.API{}
	sc.Init(apiKey, nil)

	return &Service{
		client: sc,
		r:      repo,
		appUrl: os.Getenv("APP_URL"),
		planMap: map[string]string{
			"premium": os.Getenv("PREMIUM_PRICE_ID"),
			"pro":     os.Getenv("PRO_PRICE_ID"),
		},
	}
}

func (s *Service) CreateCustomer(email, firstName, lastName string) (*string, error) {
	stripe.Key = s.client.AppsSecrets.Key
	stripeCustomer := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(firstName + " " + lastName),
	}

	result, err := customer.New(stripeCustomer)
	if err != nil {
		return nil, err
	}

	return &result.ID, nil
}

func (s *Service) CreateCheckoutSession(req *CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	priceID, ok := s.planMap[req.Plan]
	if !ok {
		if req.Plan != "premium" && req.Plan != "pro" {
			return nil, ErrMissingPriceID
		}
		return nil, ErrInvalidPriceID
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card", "revolut_pay"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{{
			Price:    stripe.String(priceID),
			Quantity: stripe.Int64(1),
		}},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(s.appUrl + "/dashboard?payout=success"),
		CancelURL:  stripe.String(s.appUrl + "/dashboard?payout=cancel"),
	}

	if req.CustomerId != "" {
		params.Customer = stripe.String(req.CustomerId)
	}

	return s.client.CheckoutSessions.New(params)
}

func (s *Service) GetSubscription(subscriptionID string) (*stripe.Subscription, error) {
	return s.client.Subscriptions.Get(subscriptionID, nil)
}

func (s *Service) CancelSubscription(subscriptionID string) (*stripe.Subscription, error) {
	return s.client.Subscriptions.Cancel(subscriptionID, nil)
}

func (s *Service) HandleWebhook(c echo.Context) error {
	ctx := c.Request().Context()
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	defer c.Request().Body.Close()

	var event stripe.Event
	if err := json.Unmarshal(body, &event); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	var handlerErr error

	switch event.Type {
	case "invoice.payment_succeeded":
		handlerErr = s.handleInvoicePaymentSucceeded(ctx, event)
	case "invoice.payment_failed":
		handlerErr = s.handleInvoicePaymentFailed(ctx, event)
	case "customer.subscription.updated":
		handlerErr = s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		handlerErr = s.handleSubscriptionDeleted(ctx, event)
	case "checkout.session.completed":
		handlerErr = s.handleCheckoutSessionCompleted(ctx, event)
	default:
		return c.NoContent(http.StatusOK)
	}

	if handlerErr != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
