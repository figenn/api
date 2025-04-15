package portfolio

import "time"

type Stock struct {
	ID            string
	Ticker        string
	Name          string
	Sector        string
	Industry      string
	Exchange      string
	Currency      string
	Country       string
	LogoURL       string
	DividendYield float64
	PERatio       float64
	MarketCap     int64
	CreatedAt     time.Time
}

type SearchResponse struct {
	BestMatches []Match `json:"bestMatches"`
}

type Match struct {
	Symbol   string `json:"1. symbol"`
	Name     string `json:"2. name"`
	Region   string `json:"4. region"`
	Currency string `json:"8. currency"`
}

type Overview struct {
	Symbol        string  `json:"Symbol"`
	Name          string  `json:"Name"`
	Sector        string  `json:"Sector"`
	Industry      string  `json:"Industry"`
	Exchange      string  `json:"Exchange"`
	Currency      string  `json:"Currency"`
	Country       string  `json:"Country"`
	MarketCap     int64   `json:"MarketCapitalization,string"`
	PERatio       float64 `json:"PERatio,string"`
	DividendYield float64 `json:"DividendYield,string"`
}

type CreateStockRequest struct {
	Name         string  `json:"name" form:"name"`
	Currency     string  `json:"currency" form:"currency"`
	Shares       float64 `json:"shares" form:"shares"`
	AvgPrice     float64 `json:"avg_price" form:"avg_price"`
	Notes        string  `json:"notes" form:"notes"`
	PurchaseDate string  `json:"purchase_date" form:"purchase_date"`
}
