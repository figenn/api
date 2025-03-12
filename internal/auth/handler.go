package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type API struct {
	service *Service
}

func NewAPI(service *Service) *API {
	return &API{
		service: service,
	}
}

func (a *API) Bind(rg *echo.Group) {
	authGroup := rg.Group("/auth")
	authGroup.POST("/register", a.Register)
	authGroup.POST("/login", a.Login)
}

func (a *API) Register(c echo.Context) error {
	var req RegisterRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Request format is invalid",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrMissingFields.Error(),
		})
	}

	resp, err := a.service.Register(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserExists):
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error: err.Error(),
			})

		case errors.Is(err, ErrInvalidEmail) || errors.Is(err, ErrPasswordTooWeak) || errors.Is(err, ErrMissingFields):
			return c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Error: err.Error(),
			})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrInternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusCreated, resp)
}

func (a *API) Login(c echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Request format is invalid",
		})
	}

	resp, err := a.service.Login(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrInvalidEmail):
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: ErrInvalidCredentials.Error(),
			})

		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrInternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}
