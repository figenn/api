package powens

import (
	"time"

	"github.com/google/uuid"
)

type WebhookPayload struct {
	WebhookReceived WebhookData `json:"webhook_received"`
}

type WebhookData struct {
	ID            int            `json:"id"`
	IDWebhookData int            `json:"id_webhook_data"`
	Platform      string         `json:"platform"`
	Signin        time.Time      `json:"signin"`
	Request       RequestDetails `json:"request_details"`
}

type RequestDetails struct {
	Time      time.Time `json:"time"`
	ID        string    `json:"id"`
	RemoteIP  string    `json:"remote_ip"`
	Host      string    `json:"host"`
	Method    string    `json:"method"`
	URI       string    `json:"uri"`
	UserAgent string    `json:"user_agent"`
	Status    int       `json:"status"`
	Error     string    `json:"error"`
	Latency   Latency   `json:"latency"`
	BytesIn   int       `json:"bytes_in"`
	BytesOut  int       `json:"bytes_out"`
}

type Latency struct {
	Microseconds  int    `json:"microseconds"`
	HumanReadable string `json:"human_readable"`
}

type PowensAccount struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	PowensID    int       `json:"powens_id"`
	AccessToken string    `json:"access_token"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreatePowensAccountRequest struct {
	UserID uuid.UUID `json:"user_id" form:"user_id"`
}

type PowensInitResponse struct {
	AuthToken string `json:"auth_token"`
	Type      string `json:"type"`
	IdUser    int    `json:"id_user"`
}

type PowensInitBody struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type TokenResponse struct {
	Token     string `json:"token"`
	Scope     string `json:"scope"`
	ExpiresIn int    `json:"expires_in"`
	ExpireIn  int    `json:"expire_in"`
}

type Transactions struct {
	Id                int     `json:"id"`
	IdAccount         int     `json:"id_account"`
	Date              string  `json:"date"`
	OriginalWording   string  `json:"original_wording"`
	SimplifiedWording string  `json:"simplified_wording"`
	Type              string  `json:"type"`
	Value             float64 `json:"value"`
	FormattedValue    string  `json:"formatted_value"`
}

type TransactionsResponse struct {
	Transactions []Transactions `json:"transactions"`
}
