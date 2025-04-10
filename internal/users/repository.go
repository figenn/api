package users

import (
	"context"
	"errors"
	"figenn/internal/database"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	s database.DbService
}

func NewRepository(db database.DbService) *Repository {
	return &Repository{
		s: db,
	}
}

func (r *Repository) GetUser(ctx context.Context, id string) (*UserRequest, error) {
	var u UserRequest

	builder, args, err := squirrel.
		Select(
			"u.id",
			"u.email",
			"u.first_name",
			"u.last_name",
			"u.profile_picture_url",
			"u.country",
			"u.created_at",
			"u.stripe_customer_id",
			"u.two_fa_enabled",
			"us.subscription_type",
			"us.status").
		From("users AS u").
		InnerJoin(`
		(
			SELECT stripe_customer_id, subscription_type, status
			FROM user_subscriptions
			WHERE status = 'active'
			ORDER BY updated_at DESC
			LIMIT 1
		) AS us ON u.stripe_customer_id = us.stripe_customer_id`).
		Where(squirrel.Eq{"u.id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}

	err = r.s.Pool().QueryRow(ctx, builder, args...).Scan(
		&u.ID,
		&u.Email,
		&u.FirstName,
		&u.LastName,
		&u.ProfilePictureUrl,
		&u.Country,
		&u.CreatedAt,
		&u.StripeCustomerID,
		&u.TwoFAEnabled,
		&u.SubscriptionType,
		&u.Status,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) GetUserByEmail(email string) (*User, error) {
	query, args, err := squirrel.
		Select("*").
		From("users").
		Where(squirrel.Eq{"email": email}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	var user User
	err = r.s.Pool().QueryRow(context.Background(), query, args...).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.IsResettingPassword,
		&user.ResetPasswordToken,
		&user.DateResetPassword,
		&user.ProfilePictureUrl,
		&user.StripeCustomerID,
		&user.Bio,
		&user.Country,
		&user.Currency,
		&user.LastLogin,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetActiveSubscriptionByCustomerID(stripeCustomerID string) (*UserSubscription, error) {
	query, args, err := squirrel.
		Select(
			"id", "stripe_customer_id", "stripe_subscription_id", "stripe_price_id",
			"subscription_type", "status", "cancel_at_period_end",
			"current_period_start", "current_period_end",
			"canceled_at", "ends_at", "created_at", "updated_at",
		).
		From("user_subscriptions").
		Where(squirrel.Eq{
			"stripe_customer_id": stripeCustomerID,
			"status":             "active",
		}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	var sub UserSubscription
	err = r.s.Pool().QueryRow(context.Background(), query, args...).Scan(
		&sub.ID,
		&sub.StripeSubscriptionID,
		&sub.StripePriceID,
		&sub.SubscriptionType,
		&sub.Status,
		&sub.CurrentPeriodEnd,
	)
	if err != nil {
		return nil, err
	}

	return &sub, nil
}
