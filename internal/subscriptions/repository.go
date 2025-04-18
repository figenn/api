package subscriptions

import (
	"context"
	"database/sql"
	"errors"
	"figenn/internal/database"
	"time"

	"github.com/Masterminds/squirrel"
)

type Repository struct {
	db database.DbService
}

func NewRepository(db database.DbService) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateSubscription(ctx context.Context, sub *Subscription) error {
	query, args, err := squirrel.Insert("subscriptions").
		Columns("user_id", "name", "category", "color", "description", "start_date", "end_date", "price", "logo_url", "billing_cycle", "is_active").
		Values(sub.UserId, sub.Name, sub.Category, sub.Color, sub.Description, sub.StartDate, sub.EndDate, sub.Price, sub.LogoUrl, sub.BillingCycle, sub.IsActive).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.New("failed to build insert query")
	}
	return r.db.Pool().QueryRow(ctx, query, args...).Scan(&sub.Id)
}

func (r *Repository) GetAllSubscriptions(ctx context.Context, userID string, limit, offset int) ([]*Subscription, error) {
	query, args, err := squirrel.Select(
		"id", "user_id", "name", "category", "color", "description", "start_date", "end_date", "price", "logo_url", "is_active", "billing_cycle",
	).
		From("subscriptions").
		Where(squirrel.Eq{"user_id": userID}).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.New("failed to build select query")
	}
	rows, err := r.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, errors.New("failed to execute query")
	}
	defer rows.Close()
	var subscriptions []*Subscription
	for rows.Next() {
		sub := new(Subscription)
		err := rows.Scan(
			&sub.Id, &sub.UserId, &sub.Name, &sub.Category, &sub.Color, &sub.Description, &sub.StartDate,
			&sub.EndDate, &sub.Price, &sub.LogoUrl, &sub.IsActive, &sub.BillingCycle,
		)
		if err != nil {
			return nil, errors.New("failed to scan row")
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, rows.Err()
}

func (r *Repository) DeleteSubscription(ctx context.Context, userID, subID string) error {
	query, args, err := squirrel.Delete("subscriptions").
		Where(squirrel.Eq{"user_id": userID, "id": subID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.New("failed to build delete query")
	}
	_, err = r.db.Pool().Exec(ctx, query, args...)
	return err
}

func (r *Repository) UpdateSubscription(ctx context.Context, userID, subID string, fields map[string]interface{}) error {
	query, args, err := squirrel.Update("subscriptions").
		SetMap(fields).
		Where(squirrel.Eq{"user_id": userID, "id": subID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errors.New("failed to build update query")
	}
	_, err = r.db.Pool().Exec(ctx, query, args...)
	return err
}

func (r *Repository) GetSubscriptionByID(ctx context.Context, userID, subID string) (*Subscription, error) {
	query, args, err := squirrel.Select(
		"id", "user_id", "name", "category", "color", "description", "start_date", "price", "logo_url", "is_active", "billing_cycle",
	).
		From("subscriptions").
		Where(squirrel.Eq{"user_id": userID, "id": subID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.New("failed to build select query")
	}
	sub := new(Subscription)
	err = r.db.Pool().QueryRow(ctx, query, args...).Scan(
		&sub.Id, &sub.UserId, &sub.Name, &sub.Category, &sub.Color, &sub.Description, &sub.StartDate, &sub.Price,
		&sub.LogoUrl, &sub.IsActive, &sub.BillingCycle,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.New("failed to execute query")
	}
	return sub, nil
}

func (r *Repository) GetSubscriptionsByCategory(ctx context.Context, userID string) ([]*SubscriptionCategoryCount, error) {
	query, args, err := squirrel.Select("category", "COUNT(*) AS count").
		From("subscriptions").
		Where(squirrel.Eq{"user_id": userID, "is_active": true}).
		GroupBy("category").
		OrderBy("count DESC").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.New("failed to build category count query")
	}
	rows, err := r.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, errors.New("failed to execute category count query")
	}
	defer rows.Close()
	var result []*SubscriptionCategoryCount
	for rows.Next() {
		item := new(SubscriptionCategoryCount)
		if err := rows.Scan(&item.Category, &item.Count); err != nil {
			return nil, errors.New("failed to scan category count row")
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) GetUpcomingSubscriptions(ctx context.Context, userID string, week int) ([]*Subscription, error) {
	query, args, err := squirrel.Select("id", "user_id", "name", "start_date", "color", "logo_url", "price").
		From("subscriptions").
		Where(squirrel.Eq{"user_id": userID, "is_active": true}).
		Where(squirrel.Expr("EXTRACT(WEEK FROM start_date) = ?", week)).
		Where(squirrel.Expr("EXTRACT(YEAR FROM start_date) = ?", time.Now().Year())).
		OrderBy("start_date ASC").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.New("failed to build select query")
	}
	rows, err := r.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, errors.New("failed to execute query")
	}
	defer rows.Close()
	var subscriptions []*Subscription
	for rows.Next() {
		sub := new(Subscription)
		err := rows.Scan(&sub.Id, &sub.UserId, &sub.Name, &sub.StartDate, &sub.Color, &sub.LogoUrl, &sub.Price)
		if err != nil {
			return nil, errors.New("failed to scan row")
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, rows.Err()
}

func (r *Repository) CalculateActiveSubscriptionPrice(ctx context.Context, userID string, year, month *int) (float64, error) {
	q := squirrel.Select("COALESCE(SUM(price), 0)").
		From("subscriptions").
		Where(squirrel.Eq{"user_id": userID, "is_active": true})

	if year != nil {
		q = q.Where(squirrel.Expr(
			"EXTRACT(YEAR FROM start_date) <= ? AND (end_date IS NULL OR EXTRACT(YEAR FROM end_date) >= ?)",
			*year, *year,
		))
	}
	if month != nil {
		q = q.Where(squirrel.Expr(
			"EXTRACT(MONTH FROM start_date) <= ? AND (end_date IS NULL OR EXTRACT(MONTH FROM end_date) >= ?)",
			*month, *month,
		))
	}

	sqlQuery, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return 0, errors.New("failed to build select query")
	}

	var total float64
	err = r.db.Pool().QueryRow(ctx, sqlQuery, args...).Scan(&total)
	if err != nil {
		return 0, errors.New("failed to execute query")
	}
	return total, nil
}

func (r *Repository) GetActiveSubscriptions(ctx context.Context, userID string, year, month int) ([]*Subscription, error) {
	query := `
		SELECT id, user_id, name, category, color, description, start_date, end_date, price,
			   logo_url, is_active, billing_cycle
		FROM subscriptions
		WHERE user_id = $1
		AND is_active = TRUE
		AND (
			(EXTRACT(YEAR FROM start_date) = $2 AND EXTRACT(MONTH FROM start_date) = $3)
			OR (end_date IS NOT NULL AND EXTRACT(YEAR FROM end_date) = $2 AND EXTRACT(MONTH FROM end_date) = $3)
			OR (billing_cycle = 'monthly' AND start_date <= DATE($2 || '-' || $3 || '-01'))
			OR (billing_cycle = 'quarterly' AND start_date <= DATE($2 || '-' || $3 || '-01') AND MOD(EXTRACT(MONTH FROM start_date) - $3 + 12 * (EXTRACT(YEAR FROM start_date) - $2), 3) = 0)
			OR (billing_cycle = 'semi_annual' AND start_date <= DATE($2 || '-' || $3 || '-01') AND MOD(EXTRACT(MONTH FROM start_date) - $3 + 12 * (EXTRACT(YEAR FROM start_date) - $2), 6) = 0)
			OR (billing_cycle = 'annual' AND start_date <= DATE($2 || '-' || $3 || '-01') AND EXTRACT(MONTH FROM start_date) = $3)
			OR (billing_cycle = 'one_time' AND EXTRACT(YEAR FROM start_date) = $2 AND EXTRACT(MONTH FROM start_date) = $3)
		)
		ORDER BY start_date ASC`

	rows, err := r.db.Pool().Query(ctx, query, userID, year, month)
	if err != nil {
		return nil, errors.New("failed to execute active subscriptions query")
	}
	defer rows.Close()

	var subscriptions []*Subscription
	for rows.Next() {
		sub := new(Subscription)
		err := rows.Scan(
			&sub.Id, &sub.UserId, &sub.Name, &sub.Category, &sub.Color,
			&sub.Description, &sub.StartDate, &sub.EndDate, &sub.Price,
			&sub.LogoUrl, &sub.IsActive, &sub.BillingCycle,
		)
		if err != nil {
			return nil, errors.New("failed to scan active subscription row")
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, rows.Err()
}
