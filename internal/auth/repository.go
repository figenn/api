package auth

import (
	"context"
	"figenn/internal/database"
	"figenn/internal/user"
	"time"
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

func (r *Repository) CreateUser(ctx context.Context, user *user.User) error {
	query := `INSERT INTO users (email, password, first_name, last_name, profile_picture_url, country) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return r.s.Pool().QueryRow(ctx, query, user.Email, user.Password, user.FirstName, user.LastName, user.ProfilePictureUrl, user.Country).Scan(&user.ID)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.s.Pool().QueryRow(ctx, "SELECT id, email, password FROM users WHERE email = $1", email).Scan(&u.ID, &u.Email, &u.Password)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &u, nil
}

func (r *Repository) SavePasswordResetToken(ctx context.Context, id int, token string) (int, string, error) {
	currentTime := time.Now()
	query := `UPDATE users 
              SET reset_password_token = $2, 
                  is_resetting_password = true, 
                  date_reset_password = $3
              WHERE id = $1 
              RETURNING id, reset_password_token`

	var returnedID int
	var returnedToken string
	err := r.s.Pool().QueryRow(ctx, query, id, token, currentTime).Scan(&returnedID, &returnedToken)
	if err != nil {
		return 0, "", err
	}

	return returnedID, returnedToken, nil
}
