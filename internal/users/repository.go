package users

import (
	"context"
	"database/sql"
	"errors"
	"figenn/internal/database"
	"fmt"
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

	err := r.s.Pool().QueryRow(ctx,
		"SELECT id, email, first_name, last_name, profile_picture_url, country, created_at FROM users WHERE id = $1",
		id).Scan(
		&u.ID,
		&u.Email,
		&u.FirstName,
		&u.LastName,
		&u.ProfilePictureUrl,
		&u.Country,
		&u.CreatedAt,
	)

	fmt.Println(err)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}
