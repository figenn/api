package portfolio

import (
	"figenn/internal/users"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type API struct {
	JWTSecret string
	s         *Service
}

func NewAPI(secret string, s *Service) *API {
	return &API{
		JWTSecret: secret,
		s:         s,
	}
}

func (a *API) Bind(rg *echo.Group) {
	g := rg.Group("/portfolio", users.CookieAuthMiddleware(a.JWTSecret))
	g.POST("/stocks", a.CreateStock)
	g.GET("/search-stocks", a.SearchStocks)
}

func (a *API) CreateStock(c echo.Context) error {
	ctx := c.Request().Context()
	userId := c.Get("user_id").(string)

	var req CreateStockRequest
	if err := c.Bind(&req); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	fmt.Println("CreateStock request:", req)

	stock, err := a.s.CreateStock(ctx, userId, req.Name, req.Currency, req.Shares, req.AvgPrice, req.Notes, req.PurchaseDate)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, echo.Map{"stock": stock})
}

func (a *API) SearchStocks(c echo.Context) error {
	ctx := c.Request().Context()
	query := c.QueryParam("query")
	if query == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	results, err := a.s.SearchStocks(ctx, query)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, echo.Map{"results": results})
}
