package powens

import (
	"context"
	"figenn/internal/database"
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

func (r *Repository) SetPowensAccount(ctx context.Context, userID uuid.UUID, powensID int, accessToken string) error {
	query := `
        INSERT INTO powens_accounts (user_id, powens_id, access_token, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $4)
        ON CONFLICT (user_id) 
        DO UPDATE SET 
            access_token = $3,
            updated_at = $4
    `
	_, err := r.s.Pool().Exec(ctx, query, userID.String(), powensID, accessToken, time.Now())
	return err
}
