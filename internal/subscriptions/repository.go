package subscriptions

import (
	"context"
	"database/sql"
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

func (r *Repository) CreateSubscription(ctx context.Context, sub *Subscription) error {
	query := `
        INSERT INTO subscriptions 
        (user_id, name, category, color, description, start_date, price, logo_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `

	var id string
	err := r.s.Pool().QueryRow(
		ctx,
		query,
		sub.UserId,
		sub.Name,
		sub.Category,
		sub.Color,
		sub.Description,
		sub.StartDate,
		sub.Price,
		sub.LogoUrl,
	).Scan(&id)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (r *Repository) GetActiveSubscriptions(ctx context.Context, userID string, year int, month int) ([]*Subscription, error) {
	query := `
        SELECT 
            id,
            user_id,
            name,	
            category,
            color,
            description,
            start_date,
            end_date,
            price,
            logo_url,
            active,
            is_recuring
        FROM subscriptions
        WHERE user_id = $1
          AND active = TRUE
          AND (
            (EXTRACT(YEAR FROM start_date) = $2 AND EXTRACT(MONTH FROM start_date) = $3) OR
            (end_date IS NOT NULL AND EXTRACT(YEAR FROM end_date) = $2 AND EXTRACT(MONTH FROM end_date) = $3) OR
            (is_recuring = TRUE AND EXTRACT(YEAR FROM start_date) <= $2 AND EXTRACT(MONTH FROM start_date) <= $3)
          )
    `

	rows, err := r.s.Pool().Query(ctx, query, userID, year, month)
	if err != nil {
		return nil, sql.ErrNoRows
	}
	defer rows.Close()

	var subs []*Subscription
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(
			&sub.Id, &sub.UserId, &sub.Name, &sub.Category,
			&sub.Color, &sub.Description, &sub.StartDate,
			&sub.EndDate, &sub.Price, &sub.LogoUrl, &sub.Active, &sub.IsRecuring,
		)
		if err != nil {
			return nil, err
		}
		subs = append(subs, &sub)
	}

	return subs, nil
}

func (r *Repository) GetAllSubscriptions(ctx context.Context, userID string, limit, offset int) ([]*Subscription, error) {
	query := `
		SELECT 
			id,
			user_id,
			name,	
			category,
			color,
			description,
			start_date,
			end_date,
			price,
			logo_url,
			active,
			is_recuring
		FROM subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
		OFFSET $3
	`

	rows, err := r.s.Pool().Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, sql.ErrNoRows
	}
	defer rows.Close()

	var subs []*Subscription
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(
			&sub.Id, &sub.UserId, &sub.Name, &sub.Category,
			&sub.Color, &sub.Description, &sub.StartDate,
			&sub.EndDate, &sub.Price, &sub.LogoUrl, &sub.Active, &sub.IsRecuring,
		)
		if err != nil {
			return nil, err
		}
		subs = append(subs, &sub)
	}

	return subs, nil
}

func (r *Repository) DeleteSubscription(ctx context.Context, userID, subID string) error {
	query := `
		DELETE FROM subscriptions
		WHERE user_id = $1 AND id = $2
	`

	_, err := r.s.Pool().Exec(ctx, query, userID, subID)
	if err != nil {
		return err
	}

	return nil
}
