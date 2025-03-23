package powens

import (
	"figenn/internal/errors"
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
	powensGroup := rg.Group("/powens")
	powensGroup.POST("/create", a.createPowensAccount)
	powensGroup.GET("/transactions", a.ListTransactions, users.JWTMiddleware(a.JWTSecret))
}

func (a *API) createPowensAccount(c echo.Context) error {
	var req CreatePowensAccountRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Invalid request payload",
			"details": err.Error(),
		})
	}

	connectURL, err := a.s.CreateAccount(c, req.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Failed to create Powens account",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"url": connectURL,
	})
}

func (a *API) ListTransactions(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return errors.NewUnauthorizedError("")
	}
	err := a.s.ListTransactions(c, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Failed to list transactions",
			"details": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "List transactions",
	})
}

// // PowensWebhook gère la réception des webhooks de Powens
// func (a *API) PowensWebhook(c echo.Context) error {
// 	var webhookData WebhookPayload

// 	// Décoder le JSON reçu
// 	if err := json.NewDecoder(c.Request().Body).Decode(&webhookData); err != nil {
// 		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid webhook payload"})
// 	}

// 	fmt.Println("Webhook received:", webhookData)

// 	// Vérifier que l'ID utilisateur est bien présent
// 	idUser := webhookData.WebhookReceived.ID
// 	if idUser == 0 {
// 		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Missing id_user in webhook"})
// 	}

// 	// Récupération du token d'accès
// 	accessToken, err := getAccessToken(idUser)
// 	if err != nil {
// 		fmt.Println("Error getting access token:", err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to get access token"})
// 	}

// 	// Récupération des transactions
// 	transactions, err := getTransactions(idUser, accessToken)
// 	if err != nil {
// 		fmt.Println("Error getting transactions:", err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to get transactions"})
// 	}

// 	return c.JSON(http.StatusOK, echo.Map{
// 		"status":       "success",
// 		"transactions": transactions,
// 	})
// }
