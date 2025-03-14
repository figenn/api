package powens

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type API struct {
	service *Service
}

func NewAPI(service *Service) *API {
	return &API{service: service}
}

func (h *API) Bind(rg *echo.Group) {
	powensGroup := rg.Group("/powens")
	powensGroup.POST("/create", h.createPowensAccount)
}

func (h *API) createPowensAccount(ctx echo.Context) error {
	var req CreatePowensAccountRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"message": "Invalid request payload",
			"details": err.Error(),
		})
	}

	connectURL, err := h.service.CreateAccount(ctx, req.UserID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Failed to create Powens account",
			"details": err.Error(),
		})
	}

	return ctx.JSON(http.StatusCreated, echo.Map{
		"url": connectURL,
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
