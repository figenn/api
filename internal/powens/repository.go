package powens

import (
	"context"
	"figenn/internal/database"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
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

func (r *Repository) GetPowensAccount(ctx context.Context, userID string) (*PowensAccount, error) {
	query, args, err := squirrel.Select("access_token", "powens_id").
		From("powens_accounts").
		Where(squirrel.Eq{"user_id": userID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var accessToken string
	var powensID int
	err = r.s.Pool().QueryRow(ctx, query, args...).Scan(&accessToken, &powensID)
	if err != nil {
		return nil, fmt.Errorf("failed to get powens account: %w", err)
	}

	return &PowensAccount{
		AccessToken: accessToken,
		PowensID:    powensID,
	}, nil
}
