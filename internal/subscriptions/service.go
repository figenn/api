package subscriptions

import (
	"context"
	"time"

	"github.com/bluele/gcache"
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
		return ErrUserIDAndSubIDRequired
	}

	subscription := &Subscription{
		UserId:       userID,
		Name:         req.Name,
		Category:     req.Category,
		Color:        req.Color,
		Description:  req.Description,
		StartDate:    *req.StartDate,
		Price:        req.Price,
		LogoUrl:      &req.LogoUrl,
		BillingCycle: req.BillingCycle,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.repo.CreateSubscription(ctx, subscription)
}

func (s *Service) ListActiveSubscriptions(ctx context.Context, userID string, year, month int) ([]*Subscription, error) {
	if userID == "" {
		return nil, ErrUserIDAndSubIDRequired
	}

	return s.repo.GetActiveSubscriptions(ctx, userID, year, month)
}

func (s *Service) GetAllSubscriptions(ctx context.Context, userID string, limit, offset int) ([]*Subscription, error) {
	if userID == "" {
		return nil, ErrUserIDAndSubIDRequired
	}

	return s.repo.GetAllSubscriptions(ctx, userID, limit, offset)
}

func (s *Service) DeleteSubscription(ctx context.Context, userID, subID string) error {
	if userID == "" || subID == "" {
		return ErrUserIDAndSubIDRequired
	}

	subscription, err := s.repo.GetSubscriptionByID(ctx, userID, subID)
	if err != nil || subscription == nil {
		return ErrSubscriptionNotFound
	}

	if subscription.UserId != userID {
		return ErrUserPermissionDenied
	}

	return s.repo.DeleteSubscription(ctx, userID, subID)
}

func (s *Service) UpdateSubscription(ctx context.Context, userID, subID string, req UpdateSubscriptionRequest) error {
	if userID == "" || subID == "" {
		return ErrUserIDAndSubIDRequired
	}

	subscription, err := s.repo.GetSubscriptionByID(ctx, userID, subID)
	if err != nil || subscription == nil {
		return ErrSubscriptionNotFound
	}

	if subscription.UserId != userID {
		return ErrUserPermissionDenied
	}

	fields := make(map[string]interface{})

	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Category != nil {
		fields["category"] = *req.Category
	}
	if req.Color != nil {
		fields["color"] = *req.Color
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.StartDate != nil {
		fields["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		fields["end_date"] = *req.EndDate
	}
	if req.Price != nil {
		fields["price"] = *req.Price
	}
	if req.IsActive != nil {
		fields["is_active"] = *req.IsActive
	}
	if req.IsRecuring != nil {
		fields["is_recuring"] = *req.IsRecuring
	}

	if len(fields) == 0 {
		return ErrNoFieldsToUpdate
	}

	return s.repo.UpdateSubscription(ctx, userID, subID, fields)
}

func (s *Service) GetSubscription(ctx context.Context, userID, subID string) (*Subscription, error) {
	if userID == "" || subID == "" {
		return nil, ErrUserIDAndSubIDRequired
	}

	subscription, err := s.repo.GetSubscriptionByID(ctx, userID, subID)
	if err != nil || subscription == nil {
		return nil, ErrSubscriptionNotFound
	}

	return subscription, nil
}
