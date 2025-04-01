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

	builder, args, err := squirrel.Select(
		"users.id",
		"users.email",
		"users.first_name",
		"users.last_name",
		"users.profile_picture_url",
		"users.country",
		"users.created_at",
		"users.stripe_customer_id",
		"us.subscription_type",
		"us.status").
		From("users").
		LeftJoin("user_subscriptions AS us ON users.stripe_customer_id = us.stripe_customer_id").
		Where(squirrel.Eq{"users.id": id}).
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
