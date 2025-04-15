package portfolio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	YAHOO_API_BASE_URL      = "https://query1.finance.yahoo.com/v8/finance/chart/"
	LOGO_DEV_API_BASE_URL   = "https://img.logo.dev/ticker/"
	ALPHAVANTAGE_SEARCH_URL = "https://www.alphavantage.co/query?function=SYMBOL_SEARCH"
)

type Client struct {
	hc     *http.Client
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{
		hc: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

// Search via AlphaVantage
func (c *Client) SearchSymbol(keyword string) (*SearchResponse, error) {
	url := fmt.Sprintf("%s&keywords=%s&apikey=%s", ALPHAVANTAGE_SEARCH_URL, keyword, c.apiKey)

	resp, err := c.hc.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if bytes.Contains(body, []byte(`"Note"`)) || bytes.Contains(body, []byte(`"Error Message"`)) {
		return nil, fmt.Errorf("AlphaVantage error: %s", string(body))
	}

	var raw struct {
		Matches []map[string]string `json:"bestMatches"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	var results []Match
	for _, m := range raw.Matches {
		results = append(results, Match{
			Symbol:   m["1. symbol"],
			Name:     m["2. name"],
			Region:   m["4. region"],
			Currency: m["8. currency"],
		})
	}

	return &SearchResponse{BestMatches: results}, nil
}

func (c *Client) GetStockOverview(symbol string) (*Overview, error) {
	url := YAHOO_API_BASE_URL + symbol + "?range=1d&interval=1d"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Symbol             string  `json:"symbol"`
					LongName           string  `json:"longName"`
					Currency           string  `json:"currency"`
					ExchangeName       string  `json:"exchangeName"`
					FullExchangeName   string  `json:"fullExchangeName"`
					RegularMarketPrice float64 `json:"regularMarketPrice"`
				} `json:"meta"`
			} `json:"result"`
		} `json:"chart"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	if len(raw.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data from Yahoo for %s", symbol)
	}

	meta := raw.Chart.Result[0].Meta

	return &Overview{
		Symbol:        meta.Symbol,
		Name:          meta.LongName,
		Exchange:      meta.FullExchangeName,
		Currency:      meta.Currency,
		MarketCap:     0,
		PERatio:       0,
		DividendYield: 0,
		CurrentPrice:  meta.RegularMarketPrice,
	}, nil
}

func (c *Client) GetCurrentPrice(symbol string) (float64, error) {
	o, err := c.GetStockOverview(symbol)
	if err != nil {
		return 0, err
	}
	return o.CurrentPrice, nil
}
