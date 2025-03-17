package subscriptions

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bluele/gcache"
)

const (
	LogoUrl = "https://api.svgl.app?search="
)

type Service struct {
	repo  *Repository
	cache gcache.Cache
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo:  repo,
		cache: gcache.New(100).LRU().Expiration(time.Minute * 5).Build(),
	}
}

func (s *Service) CreateSubscription(ctx context.Context, userID string, req CreateSubscriptionRequest) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	logo, err := fetchLogo(req.Name)
	if err != nil {
		return err
	}

	startDate := req.StartDate
	nextBillingDate := startDate.AddDate(0, 1, 0)

	subscription := &Subscription{
		UserId:          userID,
		Name:            req.Name,
		Category:        req.Category,
		Color:           req.Color,
		Description:     req.Description,
		StartDate:       *req.StartDate,
		Price:           req.Price,
		LogoUrl:         logo,
		NextBillingDate: nextBillingDate,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return s.repo.CreateSubscription(ctx, subscription)
}

func fetchLogo(name string) (string, error) {
	cleanName := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	variants := []string{
		cleanName,
		strings.ReplaceAll(cleanName, "-", ""),
		cleanName + "-logo",
	}

	for _, variant := range variants {
		if logo, err := tryGetLogo(variant); err == nil {
			return logo, nil
		}
	}

	return "", errors.New("failed to retrieve logo")
}

func tryGetLogo(name string) (string, error) {
	resp, err := http.Get(LogoUrl + name)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var logos []LogoResponse
	if err := json.Unmarshal(body, &logos); err != nil || len(logos) == 0 {
		return "", errors.New("no logo found")
	}

	return logos[0].Route, nil
}

func (s *Service) ListActiveSubscriptions(ctx context.Context, userID string, year, month int) ([]*Subscription, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	return s.repo.GetActiveSubscriptions(ctx, userID, year, month)
}

func (s *Service) GetAllSubscriptions(ctx context.Context, userID string, limit, offset int) ([]*Subscription, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	return s.repo.GetAllSubscriptions(ctx, userID, limit, offset)
}

func (s *Service) DeleteSubscription(ctx context.Context, userID, subID string) error {
	if userID == "" || subID == "" {
		return errors.New("user ID and subscription ID are required")
	}

	return s.repo.DeleteSubscription(ctx, userID, subID)
}
