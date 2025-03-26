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
		"id", "email", "first_name", "last_name", "profile_picture_url", "country", "created_at", "stripe_customer_id", "subscription").
		From("users").
		Where(squirrel.Eq{"id": id}).
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
		&u.Subscription,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) GetUserByStripeID(ctx context.Context, stripeID string) (*User, error) {
	var u User

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

func (r *Repository) UpdateUserSubscription(ctx context.Context, id string, subscriptionType SubscriptionType) error {
	builder, args, err := squirrel.Update("users").
		Set("subscription", subscriptionType).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	_, err = r.s.Pool().Exec(ctx, builder, args...)
	if err != nil {
		return err
	}

	return nil
}
