package users

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionType string

const (
	Free         SubscriptionType = "free"
	Premium      SubscriptionType = "premium"
	Professional SubscriptionType = "professional"
)

type User struct {
	ID                  uuid.UUID        `json:"id"`
	FirstName           string           `json:"first_name" form:"first_name"`
	LastName            string           `json:"last_name" form:"last_name"`
	Email               string           `json:"email" form:"email"`
	Username            string           `json:"username" form:"username"`
	Password            string           `json:"password" form:"password"`
	IsResettingPassword bool             `json:"is_resetting_password" form:"is_resetting_password"`
	ResetPasswordToken  string           `json:"reset_password_token,omitempty" form:"reset_password_token"`
	DateResetPassword   *time.Time       `json:"date_reset_password,omitempty" form:"date_reset_password"`
	ProfilePictureUrl   string           `json:"profile_picture_url,omitempty" form:"profile_picture_url"`
	StripeCustomerID    string           `json:"stripe_customer_id,omitempty" form:"stripe_customer_id"`
	Subscription        SubscriptionType `json:"subscription,omitempty" form:"subscription"`
	Bio                 string           `json:"bio,omitempty" form:"bio"`
	Country             string           `json:"country,omitempty" form:"country"`
	LastLogin           *time.Time       `json:"last_login,omitempty" form:"last_login"`
	TwoFAEnabled        bool             `json:"two_fa_enabled" form:"two_fa_enabled"`
	TwoFACode           string           `json:"two_fa_code,omitempty" form:"two_fa_code"`
	CreatedAt           time.Time        `json:"created_at" form:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at" form:"updated_at"`
}

type UserRequest struct {
	ID                uuid.UUID `json:"id"`
	FirstName         string    `json:"first_name" form:"first_name"`
	LastName          string    `json:"last_name" form:"last_name"`
	Email             string    `json:"email" form:"email"`
	Country           string    `json:"country,omitempty" form:"country"`
	ProfilePictureUrl string    `json:"profile_picture_url,omitempty" form:"profile_picture_url"`
	CreatedAt         time.Time `json:"created_at" form:"created_at"`
	StripeCustomerID  string    `json:"stripe_customer_id,omitempty" form:"stripe_customer_id"`
	Subscription      string    `json:"subscription,omitempty" form:"subscription"`
}
