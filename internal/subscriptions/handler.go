package subscriptions

import (
	"context"
	"figenn/internal/errors"
	"figenn/internal/users"
	"figenn/internal/utils"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// mockgen -source=internal/subscriptions/handler.go -destination=internal/subscriptions/mocks/mock_database.go -package=mocks
type SubscriptionStore interface {
	CreateSubscription(ctx context.Context, sub *Subscription) error
	GetActiveSubscriptions(ctx context.Context, userID string, year int, month int) ([]*Subscription, error)
	GetAllSubscriptions(ctx context.Context, userID string, limit, offset int) ([]*Subscription, error)
	DeleteSubscription(ctx context.Context, userID, subID string) error
	UpdateSubscription(ctx context.Context, userID, subID string, req UpdateSubscriptionRequest) error
	GetSubscription(ctx context.Context, userID, subID string) (*Subscription, error)
	GetSubscriptionsByCategory(ctx context.Context, userID string) ([]*Subscription, error)
	GetUpcomingSubscriptions(ctx context.Context, userID string, week int) ([]*Subscription, error)
	CalculateActiveSubscriptions(ctx context.Context, userID string, year, month *int) ([]*Subscription, error)
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
	subGroup := rg.Group("/subscriptions", users.CookieAuthMiddleware(a.JWTSecret))

	subGroup.GET("", a.GetAllSubscriptions)
	subGroup.POST("", a.CreateSubscription)
	subGroup.GET("/active", a.ListActiveSubscriptions)
	subGroup.DELETE("/:id", a.DeleteSubscription)
	subGroup.PATCH("/:id", a.UpdateSubscription)
	subGroup.GET("/:id", a.GetSubscription)
	subGroup.GET("/calculate", a.CalculateActiveSubscriptions)
	subGroup.GET("/upcoming", a.GetUpcomingSubscriptions)
	subGroup.GET("/by_category", a.GetSubscriptionsByCategory)
}

func (a *API) CreateSubscription(c echo.Context) error {
	ctx := c.Request().Context()
	var req CreateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("Invalid request format")
	}

	if !isValidBillingCycle(req.BillingCycle) {
		return errors.NewBadRequestError("Invalid billing cycle value")
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if err := a.s.CreateSubscription(ctx, userID, req); err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "Subscription created successfully"})
}

func (a *API) ListActiveSubscriptions(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	year, err := utils.ValidateYear(c.QueryParam("year"))
	if err != nil {
		return errors.NewBadRequestError(err.Error())
	}

	month, err := utils.ValidateMonth(c.QueryParam("month"))
	if err != nil {
		return errors.NewBadRequestError(err.Error())
	}

	subs, err := a.s.ListActiveSubscriptions(c.Request().Context(), userID, year, month)
	if err != nil {
		return errors.NewInternalServerError("Failed to fetch subscriptions")
	}

	return c.JSON(http.StatusOK, subs)
}

func (a *API) GetAllSubscriptions(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	limit, offset, err := utils.GetPaginationParams(c)
	if err != nil {
		return errors.NewBadRequestError(err.Error())
	}

	subs, err := a.s.GetAllSubscriptions(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return errors.NewInternalServerError("Failed to fetch subscriptions")
	}

	return c.JSON(http.StatusOK, subs)
}

func (a *API) DeleteSubscription(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	subID := c.Param("id")
	if subID == "" {
		return errors.NewBadRequestError("Invalid subscription ID")
	}

	if err := a.s.DeleteSubscription(c.Request().Context(), userID, subID); err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Subscription deleted successfully"})
}

func (a *API) UpdateSubscription(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	subID := c.Param("id")
	if subID == "" {
		return errors.NewBadRequestError("Invalid subscription ID")
	}

	var req UpdateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequestError("Invalid request format")
	}

	if err := a.s.UpdateSubscription(c.Request().Context(), userID, subID, req); err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Subscription updated successfully"})
}

func (a *API) GetSubscription(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	subID := c.Param("id")
	if subID == "" {
		return errors.NewBadRequestError("Invalid subscription ID")
	}

	sub, err := a.s.GetSubscription(c.Request().Context(), userID, subID)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, sub)
}

func (a *API) CalculateActiveSubscriptions(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	year, err := optionalIntFromQuery(c, "year", utils.ValidateYear)
	if err != nil {
		return errors.NewBadRequestError(err.Error())
	}

	month, err := optionalIntFromQuery(c, "month", utils.ValidateMonth)
	if err != nil {
		return errors.NewBadRequestError(err.Error())
	}

	subs, err := a.s.CalculateActiveSubscriptions(c.Request().Context(), userID, year, month)
	if err != nil {
		return errors.NewInternalServerError("Failed to fetch subscriptions")
	}

	return c.JSON(http.StatusOK, subs)
}

func (a *API) GetUpcomingSubscriptions(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	weekStr := c.QueryParam("week")
	if weekStr == "" {
		return errors.NewBadRequestError("Week is required")
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil {
		return errors.NewBadRequestError("Invalid week value")
	}

	subs, err := a.s.GetUpcomingSubscriptions(c.Request().Context(), userID, week)
	if err != nil {
		return errors.NewInternalServerError("Failed to fetch subscriptions")
	}

	return c.JSON(http.StatusOK, subs)
}

func (a *API) GetSubscriptionsByCategory(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	counts, err := a.s.GetSubscriptionsByCategory(c.Request().Context(), userID)
	if err != nil {
		return errors.NewInternalServerError("Failed to fetch subscription counts by category")
	}

	return c.JSON(http.StatusOK, counts)
}

func isValidBillingCycle(cycle BillingCycleType) bool {
	switch cycle {
	case Monthly, Quarterly, Annual:
		return true
	default:
		return false
	}
}

func getUserID(c echo.Context) (string, error) {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return "", errors.NewUnauthorizedError("")
	}
	return userID, nil
}

func optionalIntFromQuery(c echo.Context, key string, validate func(string) (int, error)) (*int, error) {
	val := c.QueryParam(key)
	if val == "" {
		return nil, nil
	}
	v, err := validate(val)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func handleServiceError(err error) error {
	switch err {
	case ErrUserIDAndSubIDRequired:
		return errors.NewBadRequestError("User ID and Subscription ID are required")
	case ErrSubscriptionNotFound:
		return errors.NewNotFoundError("Subscription not found")
	case ErrUserPermissionDenied:
		return errors.NewForbiddenError("You are not authorized to perform this action")
	case ErrFailedCreateSub:
		return errors.NewInternalServerError("Failed to create subscription")
	default:
		return errors.NewInternalServerError("Unexpected subscription service error")
	}
}
