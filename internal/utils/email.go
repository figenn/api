package utils

import (
	"context"
	"figenn/internal/mailer"
	"figenn/internal/users"
	"log"
)

func SendWelcomeEmail(mailerClient mailer.Mailer, user *users.User) {
	ctx := context.Background()
	emailConfig := mailer.Config{
		To:      user.Email,
		Subject: "Welcome to our application",
		Html:    "<p>Hello " + user.FirstName + ",</p><p>Thank you for signing up for our application.</p>",
	}

	_, err := mailerClient.SendMail(ctx, emailConfig)
	if err != nil {
		log.Println("Failed to send welcome email", err)
	}
}

func SendResetPasswordEmail(mailerClient mailer.Mailer, user *users.User, resetLink string) {
	ctx := context.Background()
	emailConfig := mailer.Config{
		To:      user.Email,
		Subject: "Password Reset",
		Html:    "<p>Hello " + user.FirstName + ",</p><p>Click the following link to reset your password: <a href=\"" + resetLink + "\">Reset Password</a></p>",
	}

	_, err := mailerClient.SendMail(ctx, emailConfig)
	if err != nil {
		log.Println("Failed to send reset password email", err)
	}
}
