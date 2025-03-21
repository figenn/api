package subscriptions

import (
	"context"
	"database/sql"
	"errors"
	"figenn/internal/database"

	"github.com/Masterminds/squirrel"
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
	query, args, err := squirrel.Insert("subscriptions").
		Columns("user_id", "name", "category", "color", "description", "start_date", "end_date", "price",
			"logo_url", "billing_cycle", "is_active").
		Values(sub.UserId, sub.Name, sub.Category, sub.Color, sub.Description, sub.StartDate, sub.EndDate,
			sub.Price, sub.LogoUrl, sub.BillingCycle, sub.IsActive).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return errors.New("failed to build insert query")
	}

	err = r.s.Pool().QueryRow(ctx, query, args...).Scan(&sub.Id)
	if err != nil {
		return errors.New("failed to execute insert query")
	}

	return nil
}

func (r *Repository) GetActiveSubscriptions(ctx context.Context, userID string, year, month int) ([]*Subscription, error) {
	query := `
    SELECT id, user_id, name, category, color, description, start_date, end_date, price,
           logo_url, is_active, billing_cycle
    FROM subscriptions
    WHERE user_id = $1
    AND is_active = TRUE
    AND (
        -- Abonnements qui commencent ce mois-ci
        (EXTRACT(YEAR FROM start_date) = $2 AND EXTRACT(MONTH FROM start_date) = $3)
        
        -- Abonnements qui se terminent ce mois-ci
        OR (end_date IS NOT NULL AND EXTRACT(YEAR FROM end_date) = $2 AND EXTRACT(MONTH FROM end_date) = $3)
        
        -- Abonnements mensuels 
        OR (billing_cycle = 'monthly' AND
            (start_date <= DATE($2 || '-' || $3 || '-01'))
           )
        
        -- Abonnements trimestriels 
        OR (billing_cycle = 'quarterly' AND
            (start_date <= DATE($2 || '-' || $3 || '-01')) AND
            (MOD(EXTRACT(MONTH FROM start_date) - $3 + 12 * (EXTRACT(YEAR FROM start_date) - $2), 3) = 0)
           )
        
        -- Abonnements semestriels
        OR (billing_cycle = 'semi_annual' AND
            (start_date <= DATE($2 || '-' || $3 || '-01')) AND
            (MOD(EXTRACT(MONTH FROM start_date) - $3 + 12 * (EXTRACT(YEAR FROM start_date) - $2), 6) = 0)
           )
        
        -- Abonnements annuels 
        OR (billing_cycle = 'annual' AND
            (start_date <= DATE($2 || '-' || $3 || '-01')) AND
            (EXTRACT(MONTH FROM start_date) = $3)
           )
        
        -- Abonnements one-time 
        OR (billing_cycle = 'one_time' AND
            (EXTRACT(YEAR FROM start_date) = $2 AND EXTRACT(MONTH FROM start_date) = $3)
           )
    )
    ORDER BY start_date ASC
    `

	rows, err := r.s.Pool().Query(ctx, query, userID, year, month)
	if err != nil {
		return nil, errors.New("failed to execute query")
	}
	defer rows.Close()

	var subscriptions []*Subscription
	for rows.Next() {
		var sub Subscription
		if err := rows.Scan(
			&sub.Id, &sub.UserId, &sub.Name, &sub.Category,
			&sub.Color, &sub.Description, &sub.StartDate,
			&sub.EndDate, &sub.Price, &sub.LogoUrl, &sub.IsActive,
			&sub.BillingCycle,
		); err != nil {
			return nil, errors.New("failed to scan row")
		}

		subscriptions = append(subscriptions, &sub)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("error while iterating rows")
	}

	return subscriptions, nil
}

func (r *Repository) GetAllSubscriptions(ctx context.Context, userID string, limit, offset int) ([]*Subscription, error) {
	query, args, err := squirrel.Select("id", "user_id", "name", "category", "color", "description", "start_date", "end_date", "price", "logo_url", "is_active", "billing_cycle").
		From("subscriptions").
		Where(squirrel.Eq{"user_id": userID}).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New("failed to build select query")
	}

	rows, err := r.s.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, errors.New("failed to execute query")
	}
	defer rows.Close()

	var subscriptions []*Subscription
	for rows.Next() {
		var sub Subscription
		if err := rows.Scan(
			&sub.Id, &sub.UserId, &sub.Name, &sub.Category,
			&sub.Color, &sub.Description, &sub.StartDate,
			&sub.EndDate, &sub.Price, &sub.LogoUrl, &sub.IsActive,
			&sub.BillingCycle,
		); err != nil {
			return nil, errors.New("failed to scan row")
		}

		subscriptions = append(subscriptions, &sub)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("error while iterating rows")
	}

	return subscriptions, nil
}

func (r *Repository) DeleteSubscription(ctx context.Context, userID, subID string) error {
	query, args, err := squirrel.Delete("subscriptions").
		Where(squirrel.Eq{"user_id": userID}).
		Where(squirrel.Eq{"id": subID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return errors.New("failed to build delete query")
	}

	_, err = r.s.Pool().Exec(ctx, query, args...)
	if err != nil {
		return errors.New("failed to execute delete query")
	}

	return nil
}

func (r *Repository) UpdateSubscription(ctx context.Context, userID, subID string, fields map[string]interface{}) error {
	query, args, err := squirrel.Update("subscriptions").
		SetMap(fields).
		Where(squirrel.Eq{"user_id": userID}).Where(squirrel.Eq{"id": subID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return errors.New("failed to build update query")
	}

	_, err = r.s.Pool().Exec(ctx, query, args...)
	if err != nil {
		return errors.New("failed to execute update query")
	}

	return nil
}

func (r *Repository) GetSubscriptionByID(ctx context.Context, userID, subID string) (*Subscription, error) {
	query, args, err := squirrel.Select(
		"id", "user_id", "name", "category", "color", "description",
		"start_date", "price", "logo_url", "is_active", "billing_cycle").
		From("subscriptions").
		Where(squirrel.Eq{"user_id": userID, "id": subID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New("failed to build select query")
	}

	var sub Subscription
	if err := r.s.Pool().QueryRow(ctx, query, args...).Scan(
		&sub.Id, &sub.UserId, &sub.Name, &sub.Category, &sub.Color,
		&sub.Description, &sub.StartDate, &sub.Price,
		&sub.LogoUrl, &sub.IsActive, &sub.BillingCycle,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.New("failed to execute query")
	}

	return &sub, nil
}

func (r *Repository) CalculateActiveSubscriptionPrice(ctx context.Context, userID string, year, month *int) (float64, error) {
	query := squirrel.Select("SUM(price)").From("subscriptions").Where(squirrel.Eq{"user_id": userID, "is_active": true})

	if year != nil {
		query = query.Where(squirrel.Or{
			squirrel.And{
				squirrel.LtOrEq{"EXTRACT(YEAR FROM start_date)": *year},
				squirrel.Or{
					squirrel.Eq{"end_date": nil},
					squirrel.GtOrEq{"EXTRACT(YEAR FROM end_date)": *year},
				},
			},
			squirrel.And{
				squirrel.LtOrEq{"EXTRACT(YEAR FROM start_date)": *year},
				squirrel.Eq{"EXTRACT(YEAR FROM start_date)": *year},
				squirrel.Eq{"end_date": nil},
			},
		})
	}

	if month != nil {
		query = query.Where(squirrel.Or{
			squirrel.And{
				squirrel.LtOrEq{"EXTRACT(MONTH FROM start_date)": *month},
				squirrel.Or{
					squirrel.Eq{"end_date": nil},
					squirrel.GtOrEq{"EXTRACT(MONTH FROM end_date)": *month},
				},
			},
			squirrel.And{
				squirrel.LtOrEq{"EXTRACT(MONTH FROM start_date)": *month},
				squirrel.Eq{"EXTRACT(MONTH FROM start_date)": *month},
				squirrel.Eq{"end_date": nil},
			},
		})
	}

	query = query.PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, errors.New("failed to build select query")
	}

	var total float64
	err = r.s.Pool().QueryRow(ctx, sqlQuery, args...).Scan(&total)
	if err != nil {
		return 0, errors.New("failed to execute query")
	}

	return total, nil
}
