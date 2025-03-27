package stripe

import (
	"figenn/internal/users"
	"net/http"

	"github.com/labstack/echo/v4"
)

type API struct {
	JWTSecret string
	s         *Service
}

func NewAPI(secret string, service *Service) *API {
	return &API{
		JWTSecret: secret,
		s:         service,
	}
}

func (a *API) Bind(rg *echo.Group) {
	stripeGroup := rg.Group("/stripe")
	stripeGroup.POST("/create-checkout-session", a.HandleCreateCheckoutSession, users.CookieAuthMiddleware(a.JWTSecret))
	stripeGroup.GET("/subscriptions/:id", a.HandleGetSubscription)
	stripeGroup.DELETE("/subscriptions/:id", a.HandleCancelSubscription)
	stripeGroup.POST("/webhook", a.HandleWebhook)
}

func (a *API) HandleCreateCheckoutSession(c echo.Context) error {
	var params CheckoutSessionParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Invalid request parameters",
		})
	}

	session, err := a.s.CreateCheckoutSession(&params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Failed to create checkout session",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"url": session.URL,
	})
}

func (a *API) HandleGetSubscription(c echo.Context) error {
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Subscription ID is required",
		})
	}

	subscription, err := a.s.GetSubscription(subscriptionID)
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

	subscription, err := a.s.CancelSubscription(subscriptionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to cancel subscription",
		})
	}

	return c.JSON(http.StatusOK, subscription)
}

func (a *API) HandleWebhook(c echo.Context) error {
	return a.s.HandleWebhook(c)
}
