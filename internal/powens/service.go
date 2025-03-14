package powens

import (
	"context"
	"errors"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Config struct {
	ClientID    string
	Domain      string
	CallbackURI string
}

type Service struct {
	repo   *Repository
	client *Client
	config *Config
}

func NewService(repo *Repository, client *Client, config *Config) *Service {
	return &Service{repo: repo, client: client, config: config}
}

func (s *Service) CreateAccount(ctx echo.Context, userID uuid.UUID) (*string, error) {
	authToken, powensID, err := s.client.CreatePowensAccount(ctx, userID)
	if err != nil {
		return nil, err
	}

	err = s.repo.SetPowensAccount(context.Background(), userID, powensID, authToken)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.client.CreateTemporaryToken(ctx, authToken)
	if err != nil {
		return nil, err
	}

	if s.config.ClientID == "" || s.config.Domain == "" || s.config.CallbackURI == "" {
		return nil, errors.New("missing required config values")
	}

	urlValues := url.Values{}
	urlValues.Set("domain", s.config.Domain)
	urlValues.Set("client_id", s.config.ClientID)
	urlValues.Set("redirect_uri", s.config.CallbackURI)
	urlValues.Set("code", accessToken)

	constructedURL := "https://webview.powens.com/connect?" + urlValues.Encode()

	return &constructedURL, nil
}
