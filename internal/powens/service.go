package powens

import (
	"context"
	"errors"
	"figenn/internal/subscriptions"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

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
	r      *subscriptions.Repository
}

func NewService(repo *Repository, client *Client, config *Config, subscriptionRepository *subscriptions.Repository) *Service {
	return &Service{repo: repo, client: client, config: config, r: subscriptionRepository}
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

type SubscriptionAuto struct {
	Name     string
	Date     string
	Amount   string
	Category string
}

func (s *Service) ListTransactions(ctx echo.Context, userId string) error {

	fmt.Println(userId)

	token, err := s.repo.GetPowensAccount(context.Background(), userId)
	if err != nil {
		return err
	}

	response, err := s.client.GetTransactions(ctx, 108, *token)
	if err != nil {
		fmt.Println(err)
		return err
	}

	subscriptionAutoList := []subscriptions.Subscription{}
	addedServices := make(map[string]bool)

	for _, t := range response.Transactions {
		if addedServices[t.SimplifiedWording] {
			continue
		}
		for _, sub := range SubscriptionNames {
			if t.SimplifiedWording == sub.Name {

				parsedDate, err := time.Parse("2006-01-02", t.Date)
				if err != nil {
					return err
				}

				strings.ToLower(t.SimplifiedWording)
				strings.ReplaceAll(t.SimplifiedWording, " ", "")
				// Assurez-vous que Price soit de type float64
				subscriptionAuto := subscriptions.Subscription{
					UserId:    userId,
					Name:      sub.Name,
					Price:     math.Abs(t.Value), // Conversion en valeur absolue de type float64
					Category:  sub.Category,
					StartDate: parsedDate,
				}

				subscriptionAutoList = append(subscriptionAutoList, subscriptionAuto)
				addedServices[t.SimplifiedWording] = true
				break
			}
		}
	}

	// Insert each subscription individually
	for _, subscription := range subscriptionAutoList {
		err := s.r.CreateSubscription(context.Background(), &subscription) // Pass reference to Subscription
		if err != nil {
			fmt.Println("Error while inserting subscription:", err)
			continue
		}
	}

	return nil
}
