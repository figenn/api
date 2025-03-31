package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type Client struct {
	model  string
	client *http.Client
}

func NewClient(model string) *Client {
	return &Client{
		model:  model,
		client: &http.Client{},
	}
}

func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	payload := map[string]interface{}{
		"model":    c.model,
		"messages": messages,
		"stream":   false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:11434/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.New("llm: unexpected status code")
	}

	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}

	return out.Message.Content, nil
}
