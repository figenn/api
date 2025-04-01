package payment

import (
	"context"
	"errors"
	"figenn/internal/database"
	"figenn/internal/users"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
)

type Repository struct {
	s database.DbService
}

func NewRepository(db database.DbService) *Repository {
	return &Repository{
		s: db,
	}
}

func (r *Repository) GetUserByStripeID(ctx context.Context, stripeID string) (*users.User, error) {
	var u users.User

	builder, args, err := squirrel.Select("id").
		From("users").
		Where(squirrel.Eq{"stripe_customer_id": stripeID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}

	err = r.s.Pool().QueryRow(ctx, builder, args...).Scan(&u.ID)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) UpdateUserSubscriptionFromStripeWebhook(
	ctx context.Context,
	stripeCustomerID string,
	subscriptionType SubscriptionType,
	stripeSubID string,
	stripePriceID string,
	status string,
	currentPeriodStart time.Time,
	currentPeriodEnd time.Time,
	cancelAtPeriodEnd bool,
	canceledAt *time.Time,
	endsAt *time.Time,
) error {
	updateMap := map[string]interface{}{
		"subscription_type":      subscriptionType,
		"status":                 status,
		"stripe_subscription_id": stripeSubID,
		"stripe_price_id":        stripePriceID,
		"current_period_start":   currentPeriodStart,
		"current_period_end":     currentPeriodEnd,
		"cancel_at_period_end":   cancelAtPeriodEnd,
		"canceled_at":            canceledAt,
		"ends_at":                endsAt,
	}

	query, args, err := squirrel.
		Update("user_subscriptions").
		SetMap(updateMap).
		Where(squirrel.Eq{"stripe_customer_id": stripeCustomerID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	_, err = r.s.Pool().Exec(ctx, query, args...)
	return err
}

func (r *Repository) SetSubscription(ctx context.Context, id string, to SubscriptionType, endsAt time.Time) error {
	builder, args, err := squirrel.Update("users").
		Set("subscription", to).
		Set("subscription_ends_at", endsAt).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	_, err = r.s.Pool().Exec(ctx, builder, args...)
	return err
}
