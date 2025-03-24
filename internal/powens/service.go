package powens

import (
	"context"
	"errors"
	"figenn/internal/subscriptions"
	"math"
	"net/url"
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

func (s *Service) ListTransactions(ctx echo.Context, userId string) error {
	account, err := s.repo.GetPowensAccount(ctx.Request().Context(), userId)
	if err != nil {
		return err
	}

	transactions, err := s.client.GetTransactions(ctx, account.PowensID, account.AccessToken)
	if err != nil {
		return err
	}

	groupedTransactions := groupTransactions(transactions.Transactions)
	subscriptions := s.detectSubscriptions(ctx.Request().Context(), userId, groupedTransactions)

	for _, sub := range subscriptions {
		logoURL, err := GetLogo(sub.Name)
		if err != nil {
			return err
		}
		sub.LogoUrl = &logoURL

		if err := s.r.CreateSubscription(context.Background(), &sub); err != nil {
			return err
		}
	}

	return nil
}

func groupTransactions(transactions []Transactions) map[string]map[float64][]Transactions {
	grouped := make(map[string]map[float64][]Transactions)

	for _, t := range transactions {
		wording := cleanTransactionWording(t.SimplifiedWording)
		if grouped[wording] == nil {
			grouped[wording] = make(map[float64][]Transactions)
		}
		grouped[wording][t.Value] = append(grouped[wording][t.Value], t)
	}

	return grouped
}

func (s *Service) detectSubscriptions(ctx context.Context, userId string, groupedTransactions map[string]map[float64][]Transactions) []subscriptions.Subscription {
	var subscriptionsList []subscriptions.Subscription

	for wording, priceMap := range groupedTransactions {
		for price, transactions := range priceMap {
			if len(transactions) < 2 {
				continue
			}

			sort.Slice(transactions, func(i, j int) bool {
				dateI, _ := time.Parse("2006-01-02", transactions[i].Date)
				dateJ, _ := time.Parse("2006-01-02", transactions[j].Date)
				return dateI.Before(dateJ)
			})

			date1, _ := time.Parse("2006-01-02", transactions[0].Date)
			date2, _ := time.Parse("2006-01-02", transactions[1].Date)

			if date2.Sub(date1).Hours() > (31 * 24 * 3) {
				continue
			}

			for _, sub := range SubscriptionNames {
				if matchesSubscription(wording, sub) {
					if exists, _ := s.r.SubscriptionExists(ctx, userId, sub.Name, math.Abs(price), date1); !exists {
						subscriptionsList = append(subscriptionsList, createSubscription(userId, transactions[0], sub, determineBillingCycle(date1, date2)))
					}
					break
				}
			}
		}
	}

	return subscriptionsList
}

func matchesSubscription(wording string, sub SubscriptionList) bool {
	normalizedWording := strings.ToLower(strings.ReplaceAll(wording, " ", ""))
	return sub.Regex.MatchString(wording) || levenshtein.ComputeDistance(normalizedWording[:min(10, len(normalizedWording))], strings.ToLower(sub.Name)) <= 3
}

func createSubscription(userId string, t Transactions, sub SubscriptionList, billingCycle subscriptions.BillingCycleType) subscriptions.Subscription {
	parsedDate, _ := time.Parse("2006-01-02", t.Date)
	return subscriptions.Subscription{
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

func cleanTransactionWording(wording string) string {
	normalized := strings.ToLower(strings.TrimSpace(wording))

	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")

	normalized = strings.ReplaceAll(normalized, ".", "")
	normalized = strings.Join(strings.Fields(normalized), " ")

	return normalized
}
