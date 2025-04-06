package auth

import (
	"context"
	"errors"
	"figenn/internal/payment"
	"figenn/internal/users"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) CheckUserEmailExists(ctx context.Context, email string) (bool, error) {
	q := squirrel.Select("COUNT(*)").From("users").Where(squirrel.Eq{"email": email}).PlaceholderFormat(squirrel.Dollar)
	query, args, err := q.ToSql()
	if err != nil {
		return false, err
	}
	var count int
	err = r.pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

func (r *Repository) CreateUser(ctx context.Context, user *users.User) error {
	query, args, err := squirrel.Insert("users").
		Columns("email", "password", "first_name", "last_name", "profile_picture_url", "country", "stripe_customer_id", "currency").
		Values(user.Email, user.Password, user.FirstName, user.LastName, user.ProfilePictureUrl, user.Country, user.StripeCustomerID, user.Currency).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	return r.pool.QueryRow(ctx, query, args...).Scan(&user.ID)
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*users.User, error) {
	q := squirrel.Select("id", "email", "password", "first_name", "last_name", "profile_picture_url", "country", "stripe_customer_id").
		From("users").
		Where(squirrel.Eq{"email": email}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	var u users.User
	err = r.pool.QueryRow(ctx, query, args...).Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.ProfilePictureUrl, &u.Country, &u.StripeCustomerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	return &u, err
}

func (r *Repository) SaveResetPasswordToken(ctx context.Context, id uuid.UUID, token string) (uuid.UUID, string, error) {
	q := squirrel.Update("users").
		Set("reset_password_token", token).
		Set("is_resetting_password", true).
		Set("date_reset_password", time.Now()).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING id, reset_password_token").
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return uuid.Nil, "", err
	}

	var returnedID uuid.UUID
	var returnedToken string
	err = r.pool.QueryRow(ctx, query, args...).Scan(&returnedID, &returnedToken)
	return returnedID, returnedToken, err
}

func (r *Repository) IsResetTokenValid(ctx context.Context, token string) (bool, error) {
	q := squirrel.Select("1").
		From("users").
		Where(squirrel.Eq{"reset_password_token": token}).
		Limit(1).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return false, err
	}

	var exists int
	err = r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return true, err
}

func (r *Repository) FindUserIDByResetToken(ctx context.Context, token string) (uuid.UUID, *string, bool, error) {
	q := squirrel.Select("id", "email", "date_reset_password").
		From("users").
		Where(squirrel.Eq{"reset_password_token": token, "is_resetting_password": true}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return uuid.Nil, nil, false, err
	}

	var userID uuid.UUID
	var email *string
	var dateReset time.Time
	err = r.pool.QueryRow(ctx, query, args...).Scan(&userID, &email, &dateReset)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, nil, false, nil
	}
	if time.Since(dateReset) > 24*time.Hour {
		return uuid.Nil, nil, false, nil
	}
	return userID, email, true, err
}

func (r *Repository) ClearResetToken(ctx context.Context, userID uuid.UUID) error {
	q := squirrel.Update("users").
		Set("reset_password_token", nil).
		Set("is_resetting_password", false).
		Set("date_reset_password", nil).
		Where(squirrel.Eq{"id": userID}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query, args...)
	return err
}

func (r *Repository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashed string) error {
	q := squirrel.Update("users").
		Set("password", hashed).
		Where(squirrel.Eq{"id": userID}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return err
	}

	rst, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if rst.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *Repository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string) error {
	q := squirrel.Update("users").
		Set("refresh_token", token).
		Where(squirrel.Eq{"id": userID}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, query, args...)
	return err
}

func (r *Repository) CheckRefreshToken(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
	q := squirrel.Select("1").From("users").
		Where(squirrel.Eq{"id": userID, "refresh_token": token}).
		Limit(1).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return false, err
	}

	var exists int
	err = r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return true, err
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	q := squirrel.Select("id", "email", "password", "first_name", "last_name", "profile_picture_url", "country", "stripe_customer_id").
		From("users").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	var u users.User
	err = r.pool.QueryRow(ctx, query, args...).Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.ProfilePictureUrl, &u.Country, &u.StripeCustomerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return &u, err
}

func (r *Repository) InitDefaultSubscription(ctx context.Context, stripeCustomerID string) error {
	q := squirrel.Insert("user_subscriptions").
		Columns("stripe_customer_id", "subscription_type", "status", "stripe_price_id", "stripe_subscription_id", "cancel_at_period_end", "current_period_start", "current_period_end").
		Values(stripeCustomerID, payment.Free, "active", "", "", false, time.Now(), time.Now().AddDate(0, 12, 0)).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, query, args...)
	return err
}

func (r *Repository) StoreTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	q := squirrel.Update("users").
		Set("two_fa_secret", secret).
		Where(squirrel.Eq{"id": userID}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, query, args...)
	return err
}

func (r *Repository) RetrieveTOTPSecret(ctx context.Context, userID uuid.UUID) (string, error) {
	q := squirrel.Select("two_fa_secret").From("users").Where(squirrel.Eq{"id": userID}).PlaceholderFormat(squirrel.Dollar)
	query, args, err := q.ToSql()
	if err != nil {
		return "", err
	}
	var secret string
	err = r.pool.QueryRow(ctx, query, args...).Scan(&secret)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return secret, err
}

func (r *Repository) EnableTOTP(ctx context.Context, userID uuid.UUID) error {
	q := squirrel.Update("users").Set("two_fa_enabled", true).Where(squirrel.Eq{"id": userID}).PlaceholderFormat(squirrel.Dollar)
	query, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, query, args...)
	return err
}

func (r *Repository) DisableTOTP(ctx context.Context, userID uuid.UUID) error {
	q := squirrel.Update("users").
		Set("two_fa_secret", nil).
		Set("two_fa_enabled", false).
		Where(squirrel.Eq{"id": userID}).
		PlaceholderFormat(squirrel.Dollar)
	query, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, query, args...)
	return err
}
