package portfolio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	ALPHAVANTAGE_API_BASE_URL = "https://www.alphavantage.co/query"
)

type Client struct {
	hc     *http.Client
	apikey string
}

func NewClient(apiKey string) *Client {
	return &Client{
		hc: &http.Client{
			Timeout: 30 * time.Second,
		},
		apikey: apiKey,
	}
}

func (c *Client) SearchSymbol(keyword string) (*SearchResponse, error) {
	url := ALPHAVANTAGE_API_BASE_URL + "?function=SYMBOL_SEARCH&keywords=" + keyword + "&apikey=" + c.apikey
	fmt.Println("URL:", url)
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
		return nil, fmt.Errorf("AlphaVantage rate limited or error: %s", string(body))
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
	params := url.Values{}
	params.Set("function", "OVERVIEW")
	params.Set("symbol", symbol)
	params.Set("apikey", c.apikey)

	resp, err := c.hc.Get(ALPHAVANTAGE_API_BASE_URL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var o Overview
	if err := json.NewDecoder(resp.Body).Decode(&o); err != nil {
		return nil, err
	}

	if o.Symbol == "" {
		return nil, fmt.Errorf("no overview found")
	}

	return &o, nil
}
