package subscriptions

import (
	"context"
	"figenn/internal/users"
	"figenn/internal/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type SubscriptionStore interface {
	CreateSubscription(ctx context.Context, sub *Subscription) error
	GetActiveSubscriptions(ctx context.Context, userID string, year int, month int) ([]*Subscription, error)
	GetAllSubscriptions(ctx context.Context, userID string, limit, offset int) ([]*Subscription, error)
	DeleteSubscription(ctx context.Context, userID, subID string) error
}

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
	subscriptionsGroup := rg.Group("/subscriptions", users.JWTMiddleware(a.JWTSecret))
	subscriptionsGroup.GET("", a.GetAllSubscriptions)
	subscriptionsGroup.POST("/create", a.CreateSubscription)
	subscriptionsGroup.GET("/active", a.ListActiveSubscriptions)
	subscriptionsGroup.DELETE("/:id", a.DeleteSubscription)
}

func (a *API) CreateSubscription(c echo.Context) error {
	var req CreateSubscriptionRequest
	if err := c.Bind(&req); err != nil || req.Name == "" || req.Price <= 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request format",
		})
	}

	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Invalid user session",
		})
	}

	err := a.s.CreateSubscription(c.Request().Context(), userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to create subscription",
		})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"message": "Subscription created successfully",
	})
}

func (a *API) ListActiveSubscriptions(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Invalid user session",
		})
	}

	year, err := utils.ValidateYear(c.QueryParam("year"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
	}

	month, err := utils.ValidateMonth(c.QueryParam("month"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
	}

	subs, err := a.s.ListActiveSubscriptions(c.Request().Context(), userID, year, month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to fetch subscriptions",
		})
	}

	return c.JSON(http.StatusOK, subs)
}

func (a *API) GetAllSubscriptions(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Invalid user session",
		})
	}

	limit, offset, err := utils.GetPaginationParams(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
	}

	subs, err := a.s.GetAllSubscriptions(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to fetch subscriptions",
		})
	}

	return c.JSON(http.StatusOK, subs)
}

func (a *API) DeleteSubscription(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Invalid user session",
		})
	}

	subID := c.Param("id")
	if subID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid subscription ID",
		})
	}

	err := a.s.DeleteSubscription(c.Request().Context(), userID, subID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to delete subscription",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Subscription deleted successfully",
	})
}
