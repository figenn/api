package auth

import (
	"context"
	"errors"
	"figenn/internal/users"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *users.User) error
	GetUser(ctx context.Context, cond squirrel.Sqlizer) (*users.User, error)
	UserExists(ctx context.Context, cond squirrel.Sqlizer) (bool, error)
	SaveRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string) error
	VerifyRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string) (bool, error)
	InvalidateRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string) error
	SavePasswordResetToken(ctx context.Context, userID uuid.UUID, token string) (uuid.UUID, string, error)
	ValidateResetToken(ctx context.Context, token string) error
	GetUserIDByResetToken(ctx context.Context, token string) (uuid.UUID, bool, error)
	ResetPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	ClearResetToken(ctx context.Context, userID uuid.UUID) error
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	q := squirrel.Select("COUNT(*)").From("users").Where(squirrel.Eq{"email": email}).PlaceholderFormat(squirrel.Dollar)
	query, args, err := q.ToSql()
	if err != nil {
		return false, errors.New("failed to build select query")
	}
	var count int
	err = r.pool.QueryRow(ctx, query, args...).Scan(&count)
	return count > 0, err
}

func (r *Repository) CreateUser(ctx context.Context, user *users.User) error {
	query, args, err := squirrel.Insert("users").
		Columns("email", "password", "first_name", "last_name", "profile_picture_url", "country", "subscription", "stripe_customer_id").
		Values(user.Email, user.Password, user.FirstName, user.LastName, user.ProfilePictureUrl, user.Country, user.Subscription, user.StripeCustomerID).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return errors.New("failed to build insert query")
	}

	if err := r.pool.QueryRow(ctx, query, args...).Scan(&user.ID); err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*users.User, error) {
	q := squirrel.Select("id", "email", "password", "first_name", "last_name", "profile_picture_url", "country", "subscription", "stripe_customer_id").
		From("users").
		Where(squirrel.Eq{"email": email}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return nil, errors.New("failed to build select query")
	}

	var u users.User
	err = r.pool.QueryRow(ctx, query, args...).Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.ProfilePictureUrl, &u.Country, &u.Subscription, &u.StripeCustomerID)
	if errors.Is(err, errors.New("pgx: no rows in result")) {
		return nil, ErrUserNotFound
	}

	return &u, nil
}

func (r *Repository) SavePasswordResetToken(ctx context.Context, id uuid.UUID, token string) (uuid.UUID, string, error) {
	q := squirrel.Update("users").
		Set("reset_password_token", token).
		Set("is_resetting_password", true).
		Set("date_reset_password", time.Now()).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING id, reset_password_token").
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return uuid.Nil, "", errors.New("failed to build update query")
	}

	var returnedID uuid.UUID
	var returnedToken string
	err = r.pool.QueryRow(ctx, query, args...).Scan(&returnedID, &returnedToken)
	return returnedID, returnedToken, err
}

func (r *Repository) IsValidResetToken(ctx context.Context, token string) (bool, error) {
	q := squirrel.
		Select("1").
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
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *Repository) GetUserIDByResetToken(ctx context.Context, token string) (uuid.UUID, *string, bool, error) {
	q := squirrel.Select("id", "email", "date_reset_password").
		From("users").
		Where(squirrel.Eq{"reset_password_token": token}).
		Where(squirrel.Eq{"is_resetting_password": true}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return uuid.Nil, nil, false, errors.New("failed to build select query")
	}

	var userID uuid.UUID
	var email *string
	var dateResetPassword time.Time
	err = r.pool.QueryRow(ctx, query, args...).Scan(&userID, &email, &dateResetPassword)
	if err != nil {
		if errors.Is(err, errors.New("pgx: no rows in result")) {
			return uuid.Nil, nil, false, nil
		}
		return uuid.Nil, nil, false, err
	}

	if time.Since(dateResetPassword) > 24*time.Hour {
		return uuid.Nil, nil, false, nil
	}

	return userID, email, true, nil
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
		return errors.New("failed to build update query")
	}

	_, err = r.pool.Exec(ctx, query, args...)
	return err
}

func (r *Repository) ResetPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	q := squirrel.Update("users").
		Set("password", hashedPassword).
		Where(squirrel.Eq{"id": userID}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return errors.New("failed to build update query")
	}

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *Repository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, token string) error {
	q := squirrel.Update("users").
		Set("refresh_token", token).
		Where(squirrel.Eq{"id": userID}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return errors.New("failed to build update query")
	}

	_, err = r.pool.Exec(ctx, query, args...)
	fmt.Println("SaveRefreshToken", err)
	return err
}

func (r *Repository) VerifyRefreshToken(ctx context.Context, userId uuid.UUID, token string) (bool, error) {
	q := squirrel.Select("id").From("users").
		Where(squirrel.Eq{"id": userId, "refresh_token": token}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return false, errors.New("failed to build select query")
	}

	var id uuid.UUID
	err = r.pool.QueryRow(ctx, query, args...).Scan(&id)
	if errors.Is(err, errors.New("pgx: no rows in result")) {
		return false, nil
	}
	return true, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	q := squirrel.Select("id", "email", "password", "first_name", "last_name", "profile_picture_url", "country", "subscription", "stripe_customer_id").
		From("users").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := q.ToSql()
	if err != nil {
		return nil, errors.New("failed to build select query")
	}

	var u users.User
	err = r.pool.QueryRow(ctx, query, args...).Scan(&u.ID, &u.Email, &u.Password, &u.FirstName, &u.LastName, &u.ProfilePictureUrl, &u.Country, &u.Subscription, &u.StripeCustomerID)
	if errors.Is(err, errors.New("pgx: no rows in result")) {
		return nil, ErrUserNotFound
	}

	return &u, nil
}
