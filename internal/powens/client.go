package powens

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const (
	PowensAPIBaseURL  = "https://figenn-sandbox.biapi.pro/2.0"
	EndpointAuthInit  = "/auth/init"
	EndpointAuthToken = "/auth/token"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient   HTTPClient
	clientID     string
	clientSecret string
}

func NewClient(clientID, clientSecret string, httpClient HTTPClient) *Client {
	return &Client{
		httpClient:   httpClient,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (c *Client) CreatePowensAccount(ctx echo.Context, userID uuid.UUID) (string, int, error) {
	reqBody := PowensInitBody{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
	}

	var respData PowensInitResponse
	err := c.makeRequest(ctx.Request().Context(), http.MethodPost, PowensAPIBaseURL+EndpointAuthInit, reqBody, "", &respData)
	if err != nil {
		return "", 0, errors.WithStack(err)
	}

	return respData.AuthToken, respData.IdUser, nil
}

func (c *Client) CreateTemporaryToken(ctx echo.Context, authToken string) (string, error) {
	reqBody := map[string]interface{}{"duration": 3600}

	var respData TokenResponse
	err := c.makeRequest(ctx.Request().Context(), http.MethodPost, PowensAPIBaseURL+EndpointAuthToken, reqBody, authToken, &respData)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return respData.Token, nil
}

func (c *Client) GetTransactions(ctx echo.Context, idUser int, accessToken string) (*TransactionsResponse, error) {
	var transactions TransactionsResponse

	url := PowensAPIBaseURL + "/users/" + strconv.Itoa(idUser) + "/transactions?limit=30"
	err := c.makeRequest(ctx.Request().Context(), http.MethodGet, url, nil, accessToken, &transactions)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &transactions, nil
}

func (c *Client) makeRequest(ctx context.Context, method, url string, requestBody interface{}, authToken string, responseData interface{}) error {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return errors.WithStack(err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.WithStack(err)
	}

	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(responseData); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
