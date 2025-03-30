package payment

import (
	"encoding/json"
	"figenn/internal/users"
	"fmt"
	"io"
	"log"
	"net/http"

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
	client *client.API
	r      *users.Repository
}

func NewService(apiKey string, repository *users.Repository) *Service {
	sc := &client.API{}
	sc.Init(apiKey, nil)

	return &Service{
		client: sc,
		r:      repository,
	}
}

func (s *Service) CreateCustomer(email, firstName, lastName string) (*string, error) {
	stripe.Key = s.client.AppsSecrets.Key
	stripeCustomer := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(fmt.Sprintf("%s %s", firstName, lastName)),
	}

	result, err := customer.New(stripeCustomer)
	if err != nil {
		return nil, err
	}

	return &result.ID, nil
}

func (s *Service) CreateCheckoutSession(params *CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	sessionParams := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card", "revolut_pay", "paypal"}),
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

	if params.CustomerId != "" {
		sessionParams.Customer = stripe.String(params.CustomerId)
	}

	return s.client.CheckoutSessions.New(sessionParams)
}

func (s *Service) GetSubscription(subscriptionID string) (*stripe.Subscription, error) {
	return s.client.Subscriptions.Get(subscriptionID, nil)
}

func (s *Service) CancelSubscription(subscriptionID string) (*stripe.Subscription, error) {
	return s.client.Subscriptions.Cancel(subscriptionID, nil)
}

func (s *Service) HandleWebhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)

	if err != nil {
		return c.String(http.StatusInternalServerError, "Error reading body")
	}
	defer c.Request().Body.Close()

	event := stripe.Event{}
	if err := json.Unmarshal(body, &event); err != nil {
		return c.String(http.StatusBadRequest, "Invalid JSON")
	}

	switch event.Type {
	case "invoice.payment_succeeded":
		log.Println("Invoice payment succeeded")

		// Vous pouvez obtenir les informations sur la facture et l'abonnement ici
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid JSON")
		}

		// Vérifier si l'utilisateur est associé à cette facture et mettre à jour son abonnement
		user, err := s.r.GetUserByStripeID(c.Request().Context(), invoice.Customer.ID)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "Error getting user")
		}

		// Exemple : Mettre à jour l'abonnement de l'utilisateur
		err = s.r.UpdateUserSubscription(c.Request().Context(), user.ID.String(), users.Professional)
		if err != nil {
			fmt.Println(err)
			return c.String(http.StatusInternalServerError, "Error updating user subscription")
		}

		return c.String(http.StatusOK, "Subscription updated")

	case "checkout.session.completed":
		var session stripe.CheckoutSession

		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid JSON")
		}

		// fmt.Println("Session ID:", session)

		// user, err := s.r.GetUserByStripeID(c.Request().Context(), session.Customer.ID)
		// if err != nil {
		// 	return c.String(http.StatusInternalServerError, "Error getting user")
		// }

		// if user.Subscription != users.Free {
		// 	return c.String(http.StatusOK, "User already has a subscription")
		// }

		// err = s.r.UpdateUserSubscription(c.Request().Context(), user.ID.String(), users.Premium)
		// if err != nil {
		// 	return c.String(http.StatusInternalServerError, "Error updating user subscription")
		// }

		// return c.String(http.StatusOK, "Subscription updated")
	}

	return c.String(http.StatusOK, "Webhook received")
}
