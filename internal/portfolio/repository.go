package portfolio

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db squirrel.StatementBuilderType
	pg *pgxpool.Pool
}

func NewRepository(pg *pgxpool.Pool) *Repository {
	return &Repository{
		db: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		pg: pg,
	}
}

func (r *Repository) InsertStock(ctx context.Context, s *Stock) error {
	q, args, err := r.db.Insert("stocks").
		Columns(
			"ticker", "name", "sector", "industry", "exchange",
			"currency", "country", "logo_url",
			"dividend_yield", "pe_ratio", "market_cap",
		).
		Values(
			s.Ticker, s.Name, s.Sector, s.Industry, s.Exchange,
			s.Currency, s.Country, s.LogoURL,
			s.DividendYield, s.PERatio, s.MarketCap,
		).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return err
	}

	return r.pg.QueryRow(ctx, q, args...).Scan(&s.ID, &s.CreatedAt)
}

func (r *Repository) InsertUserStock(
	ctx context.Context,
	userID string,
	stockID string,
	shares float64,
	avgPrice float64,
	notes string,
	purchaseDate string,
) error {
	q, args, err := r.db.Insert("user_stocks").
		Columns(
			"user_id", "stock_id", "shares",
			"avg_price", "notes", "purchase_date",
		).
		Values(
			userID, stockID, shares,
			avgPrice, notes, purchaseDate,
		).
		Suffix("ON CONFLICT (user_id, stock_id) DO UPDATE SET shares = EXCLUDED.shares, avg_price = EXCLUDED.avg_price, notes = EXCLUDED.notes, purchase_date = EXCLUDED.purchase_date, updated_at = CURRENT_TIMESTAMP").
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pg.Exec(ctx, q, args...)
	return err
}

func (r *Repository) GetStockByTicker(ctx context.Context, ticker string) (*Stock, error) {
	q, args, err := r.db.Select(
		"id", "ticker", "name", "sector", "industry", "exchange",
		"currency", "country", "logo_url",
		"dividend_yield", "pe_ratio", "market_cap", "created_at",
	).
		From("stocks").
		Where(squirrel.Eq{"ticker": ticker}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	var s Stock
	err = r.pg.QueryRow(ctx, q, args...).Scan(
		&s.ID, &s.Ticker, &s.Name, &s.Sector, &s.Industry, &s.Exchange,
		&s.Currency, &s.Country, &s.LogoURL,
		&s.DividendYield, &s.PERatio, &s.MarketCap, &s.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return &s, err
}
