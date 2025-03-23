package powens

import (
	"context"
	"errors"
	"figenn/internal/subscriptions"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
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

	response, err := s.client.GetTransactions(ctx, token.PowensID, token.AccessToken)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Organize transactions by wording and price
	transactionsByWordingAndPrice := make(map[string]map[float64][]Transactions)
	for _, t := range response.Transactions {
		cleanedWording := cleanTransactionWording(t.SimplifiedWording)
		if transactionsByWordingAndPrice[cleanedWording] == nil {
			transactionsByWordingAndPrice[cleanedWording] = make(map[float64][]Transactions)
		}
		transactionsByWordingAndPrice[cleanedWording][t.Value] = append(transactionsByWordingAndPrice[cleanedWording][t.Value], t)
	}

	subscriptionAutoList := []subscriptions.Subscription{}

	for wording, priceMap := range transactionsByWordingAndPrice {
		for price, transactions := range priceMap {
			normalizedWording := strings.ToLower(wording)
			normalizedWording = strings.ReplaceAll(normalizedWording, " ", "")

			if len(transactions) >= 2 {
				sort.Slice(transactions, func(i, j int) bool {
					dateI, _ := time.Parse("2006-01-02", transactions[i].Date)
					dateJ, _ := time.Parse("2006-01-02", transactions[j].Date)
					return dateI.Before(dateJ)
				})

				date1, _ := time.Parse("2006-01-02", transactions[0].Date)
				date2, _ := time.Parse("2006-01-02", transactions[1].Date)

				if date2.Sub(date1).Hours() <= (31 * 24 * 3) {
					for _, sub := range SubscriptionNames {
						if sub.Regex.MatchString(wording) || levenshtein.ComputeDistance(normalizedWording[:min(10, len(normalizedWording))], strings.ToLower(sub.Name)) <= 3 {
							billingCycle := determineBillingCycle(date1, date2)

							exists, err := s.r.SubscriptionExists(ctx.Request().Context(), userId, sub.Name, math.Abs(price), date1)
							if err != nil {
								fmt.Println("Error checking subscription existence:", err)
								continue // Passer à la transaction suivante en cas d'erreur
							}

							if !exists {
								addSubscription(userId, transactions[0], sub, &subscriptionAutoList, billingCycle)
							}
							break
						}
					}
				}
			}
		}
	}

	for _, subscription := range subscriptionAutoList {
		err := s.r.CreateSubscription(context.Background(), &subscription)
		if err != nil {
			fmt.Println("Error while inserting subscription:", err)
			continue
		}
	}

	return nil
}

func addSubscription(userId string, t Transactions, sub SubscriptionList, subscriptionAutoList *[]subscriptions.Subscription, billingCycle subscriptions.BillingCycleType) {
	parsedDate, err := time.Parse("2006-01-02", t.Date)
	if err != nil {
		return
	}

	subscriptionAuto := subscriptions.Subscription{
		UserId:       userId,
		Name:         sub.Name,
		Price:        math.Abs(t.Value),
		Category:     sub.Category,
		StartDate:    parsedDate,
		BillingCycle: billingCycle,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	*subscriptionAutoList = append(*subscriptionAutoList, subscriptionAuto)
}

func cleanTransactionWording(wording string) string {
	wording = regexp.MustCompile(`(?i)(PAYPAL|CARD|CB|\d{4,})`).ReplaceAllString(wording, "")
	wording = regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(wording, "")
	return strings.TrimSpace(wording)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func determineBillingCycle(date1, date2 time.Time) subscriptions.BillingCycleType {
	diff := date2.Sub(date1).Hours()
	if diff <= (31 * 24) {
		return subscriptions.Monthly
	} else if diff <= (365 * 24) {
		return subscriptions.Annual
	}
	return subscriptions.Quarterly
}
