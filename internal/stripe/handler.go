package stripe

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v81"
)

type StripeService interface {
	CreateCheckoutSession(params *CheckoutSessionParams) (*stripe.CheckoutSession, error)
	GetSubscription(subscriptionID string) (*stripe.Subscription, error)
	CancelSubscription(subscriptionID string) (*stripe.Subscription, error)
}

type API struct {
	service StripeService
}

func NewAPI(service StripeService) *API {
	return &API{
		service: service,
	}
}

func (a *API) Bind(rg *echo.Group) {
	stripeGroup := rg.Group("/stripe")
	stripeGroup.POST("/create-checkout-session", a.HandleCreateCheckoutSession)
	stripeGroup.GET("/subscriptions/:id", a.HandleGetSubscription)
	stripeGroup.DELETE("/subscriptions/:id", a.HandleCancelSubscription)
}

func (a *API) HandleCreateCheckoutSession(c echo.Context) error {
	var params CheckoutSessionParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request parameters",
		})
	}

	session, err := a.service.CreateCheckoutSession(&params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create checkout session",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"session_id": session.ID,
		"url":        session.URL,
	})
}

func (a *API) HandleGetSubscription(c echo.Context) error {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Subscription ID is required",
		})
	}

	subscription, err := a.service.GetSubscription(subscriptionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get subscription",
		})
	}

	return c.JSON(http.StatusOK, subscription)
}

func (a *API) HandleCancelSubscription(c echo.Context) error {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Subscription ID is required",
		})
	}

	subscription, err := a.service.CancelSubscription(subscriptionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to cancel subscription",
		})
	}

	return c.JSON(http.StatusOK, subscription)
}
