package subscriptions

import (
	"context"
	"errors"
	"time"
)

type Service struct {
	r *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		r: repo,
	}
}

func (s *Service) CreateSubscription(ctx context.Context, userID string, req CreateSubscriptionRequest) error {
	sub := &Subscription{
		UserId:       userID,
		Name:         req.Name,
		Category:     req.Category,
		Color:        req.Color,
		Description:  req.Description,
		StartDate:    time.Now(),
		EndDate:      req.EndDate,
		Price:        req.Price,
		LogoUrl:      &req.LogoUrl,
		BillingCycle: req.BillingCycle,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	return s.r.CreateSubscription(ctx, sub)
}

func (r *Repository) GetActiveSubscriptions(ctx context.Context, userID string, year, month int) ([]*Subscription, error) {
	query := `
		SELECT id, user_id, name, category, color, description, start_date, end_date, price,
			   logo_url, is_active, billing_cycle
		FROM subscriptions
		WHERE user_id = $1
		AND is_active = TRUE
		AND (
			(EXTRACT(YEAR FROM start_date) = $2 AND EXTRACT(MONTH FROM start_date) = $3)
			OR (end_date IS NOT NULL AND EXTRACT(YEAR FROM end_date) = $2 AND EXTRACT(MONTH FROM end_date) = $3)
			OR (billing_cycle = 'monthly' AND start_date <= DATE($2 || '-' || $3 || '-01'))
			OR (billing_cycle = 'quarterly' AND start_date <= DATE($2 || '-' || $3 || '-01') AND MOD(EXTRACT(MONTH FROM start_date) - $3 + 12 * (EXTRACT(YEAR FROM start_date) - $2), 3) = 0)
			OR (billing_cycle = 'semi_annual' AND start_date <= DATE($2 || '-' || $3 || '-01') AND MOD(EXTRACT(MONTH FROM start_date) - $3 + 12 * (EXTRACT(YEAR FROM start_date) - $2), 6) = 0)
			OR (billing_cycle = 'annual' AND start_date <= DATE($2 || '-' || $3 || '-01') AND EXTRACT(MONTH FROM start_date) = $3)
			OR (billing_cycle = 'one_time' AND EXTRACT(YEAR FROM start_date) = $2 AND EXTRACT(MONTH FROM start_date) = $3)
		)
		ORDER BY start_date ASC`

	rows, err := r.db.Pool().Query(ctx, query, userID, year, month)
	if err != nil {
		return nil, errors.New("failed to execute active subscriptions query")
	}
	defer rows.Close()

	var subscriptions []*Subscription
	for rows.Next() {
		sub := new(Subscription)
		err := rows.Scan(
			&sub.Id, &sub.UserId, &sub.Name, &sub.Category, &sub.Color,
			&sub.Description, &sub.StartDate, &sub.EndDate, &sub.Price,
			&sub.LogoUrl, &sub.IsActive, &sub.BillingCycle,
		)
		if err != nil {
			return nil, errors.New("failed to scan active subscription row")
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, rows.Err()
}

func (s *Service) ListActiveSubscriptions(ctx context.Context, userID string, year, month int) ([]*Subscription, error) {
	if userID == "" {
		return nil, ErrUserIDAndSubIDRequired
	}
	return s.r.GetActiveSubscriptions(ctx, userID, year, month)
}

func (s *Service) GetAllSubscriptions(ctx context.Context, userID string, limit, offset int) ([]*Subscription, error) {
	return s.r.GetAllSubscriptions(ctx, userID, limit, offset)
}

func (s *Service) DeleteSubscription(ctx context.Context, userID, subID string) error {
	if userID == "" || subID == "" {
		return ErrUserIDAndSubIDRequired
	}
	return s.r.DeleteSubscription(ctx, userID, subID)
}

func (s *Service) UpdateSubscription(ctx context.Context, userID, subID string, req UpdateSubscriptionRequest) error {
	if userID == "" || subID == "" {
		return ErrUserIDAndSubIDRequired
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

	fields["updated_at"] = time.Now()

	return s.r.UpdateSubscription(ctx, userID, subID, fields)
}

func (s *Service) GetSubscription(ctx context.Context, userID, subID string) (*Subscription, error) {
	if userID == "" || subID == "" {
		return nil, ErrUserIDAndSubIDRequired
	}
	return s.r.GetSubscriptionByID(ctx, userID, subID)
}

func (s *Service) CalculateActiveSubscriptions(ctx context.Context, userID string, year, month *int) (float64, error) {
	return s.r.CalculateActiveSubscriptionPrice(ctx, userID, year, month)
}

func (s *Service) GetUpcomingSubscriptions(ctx context.Context, userID string, week int) ([]*Subscription, error) {
	return s.r.GetUpcomingSubscriptions(ctx, userID, week)
}

func (s *Service) GetSubscriptionsByCategory(ctx context.Context, userID string) ([]*SubscriptionCategoryCount, error) {
	return s.r.GetSubscriptionsByCategory(ctx, userID)
}
