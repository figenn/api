package portfolio

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bluele/gcache"
)

type Service struct {
	repo   *Repository
	client *Client
	cache  gcache.Cache
}

func NewService(repo *Repository, client *Client) *Service {
	return &Service{
		repo:   repo,
		client: client,
		cache:  gcache.New(100).LRU().Expiration(5 * time.Minute).Build(),
	}
}

func (s *Service) CreateStock(ctx context.Context, userId, name, currency string, shares, avgPrice float64, notes, purchaseDate string) (*Stock, error) {
	fmt.Println("▶️ CreateStock called")
	fmt.Println("➡️ Inputs:", name, currency, shares, avgPrice, notes, purchaseDate)

	overview, err := s.client.GetStockOverview(name)
	if err != nil {
		fmt.Println("❌ Error getting overview:", err)
		return nil, ErrOverviewNotFound
	}

	logo := "https://img.logo.dev/ticker/" + overview.Symbol + "?token=" + os.Getenv("LOGO_DEV_API_KEY")

	stock := &Stock{
		Ticker:        overview.Symbol,
		Name:          overview.Name,
		Sector:        overview.Sector,
		Industry:      overview.Industry,
		Exchange:      overview.Exchange,
		Currency:      currency,
		LogoURL:       logo,
		Country:       overview.Country,
		DividendYield: overview.DividendYield,
		PERatio:       overview.PERatio,
		MarketCap:     overview.MarketCap,
		CurrentPrice:  overview.CurrentPrice,
	}

	if err := s.repo.InsertStock(ctx, stock); err != nil {
		return nil, err
	}

	if err := s.repo.InsertUserStock(ctx, userId, stock.ID, shares, avgPrice, notes, purchaseDate); err != nil {
		return nil, err
	}

	return stock, nil
}

func (s *Service) SearchStocks(ctx context.Context, query string) ([]Match, error) {
	if v, err := s.cache.GetIFPresent(query); err == nil {
		return v.([]Match), nil
	}

	result, err := s.client.SearchSymbol(query)
	if err != nil {
		return nil, err
	}
	fmt.Println("Search result:", result)

	s.cache.SetWithExpire(query, result.BestMatches, 5*time.Minute)
	return result.BestMatches, nil
}
