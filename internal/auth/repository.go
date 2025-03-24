package auth

import (
	"context"
	"database/sql"
	"errors"
	"figenn/internal/database"
	"figenn/internal/users"
	user "figenn/internal/users"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	s database.DbService
}

func NewRepository(db database.DbService) *Repository {
	return &Repository{
		s: db,
	}
}

func (r *Repository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int
	err := r.s.Pool().QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Repository) CreateUser(ctx context.Context, user *users.User) error {
	query := `INSERT INTO users (email, password, first_name, last_name, profile_picture_url, country, subscription, stripe_customer_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	return r.s.Pool().QueryRow(ctx, query, user.Email, user.Password, user.FirstName, user.LastName, user.ProfilePictureUrl, user.Country, user.Subscription, user.StripeCustomerID).Scan(&user.ID)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.s.Pool().QueryRow(ctx, "SELECT id, email, password FROM users WHERE email = $1", email).Scan(&u.ID, &u.Email, &u.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	return &u, nil
}

func (r *Repository) SavePasswordResetToken(ctx context.Context, id uuid.UUID, token string) (uuid.UUID, string, error) {
	currentTime := time.Now()
	query := `UPDATE users 
              SET reset_password_token = $2, 
                  is_resetting_password = true, 
                  date_reset_password = $3
              WHERE id = $1 
              RETURNING id, reset_password_token`

	var returnedID uuid.UUID
	var returnedToken string
	err := r.s.Pool().QueryRow(ctx, query, id, token, currentTime).Scan(&returnedID, &returnedToken)
	if err != nil {
		return uuid.Nil, "", err

	}

	return returnedID, returnedToken, nil
}

func (r *Repository) ValidateResetToken(ctx context.Context, token string) error {
	var id int
	err := r.s.Pool().QueryRow(ctx, "SELECT id FROM users WHERE reset_password_token = $1", token).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidToken
	}

	return nil
}

func (r *Repository) GetUserIDByResetToken(ctx context.Context, token string) (uuid.UUID, bool, error) {
	var userID uuid.UUID
	var dateResetPassword time.Time

	query := `SELECT id, date_reset_password 
              FROM users 
              WHERE reset_password_token = $1 
              AND is_resetting_password = true`

	err := r.s.Pool().QueryRow(ctx, query, token).Scan(&userID, &dateResetPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}

	if time.Since(dateResetPassword) > 24*time.Hour {
		return uuid.Nil, false, nil
	}

	return userID, true, nil
}

func (r *Repository) ClearResetToken(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users 
              SET reset_password_token = NULL, 
                  is_resetting_password = false, 
                  date_reset_password = NULL 
              WHERE id = $1`

	_, err := r.s.Pool().Exec(ctx, query, userID)
	return err
}

func (r *Repository) ResetPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	query := `UPDATE users SET password = $2 WHERE id = $1`

	result, err := r.s.Pool().Exec(ctx, query, userID, hashedPassword)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
