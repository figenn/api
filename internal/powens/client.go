package powens

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const (
	PowensAPIBaseURL  = "https://figenn-sandbox.biapi.pro/2.0"
	EndpointAuthInit  = "/auth/init"
	EndpointAuthToken = "/auth/token"
)

type Client struct {
	hc           *http.Client
	clientID     string
	clientSecret string
}

func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		hc: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
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
	err := c.doRequest(ctx, http.MethodPost, PowensAPIBaseURL+EndpointAuthInit, reqBody, "", &respData)
	if err != nil {
		return "", 0, errors.WithStack(err)
	}

	return respData.AuthToken, respData.IdUser, nil
}

func (c *Client) CreateTemporaryToken(ctx echo.Context, authToken string) (string, error) {
	reqBody := map[string]interface{}{"duration": 3600}

	var respData TokenResponse
	err := c.doRequest(ctx, http.MethodPost, PowensAPIBaseURL+EndpointAuthToken, reqBody, authToken, &respData)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return respData.Token, nil
}

func (c *Client) doRequest(ctx echo.Context, method, url string, requestBody interface{}, authToken string, responseData interface{}) error {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return errors.WithStack(err)
	}

	req, err := http.NewRequestWithContext(ctx.Request().Context(), method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.WithStack(err)
	}

	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	resp, err := c.hc.Do(req)
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
