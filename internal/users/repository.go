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
