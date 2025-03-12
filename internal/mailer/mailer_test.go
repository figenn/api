package mailer_test

import (
	"context"
	"figenn/internal/mailer"
	"figenn/internal/mailer/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSendMailSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMailer := mocks.NewMockMailer(ctrl)

	ctx := context.Background()
	emailConfig := mailer.Config{
		From:    "test@example.com",
		To:      "recipient@example.com",
		Subject: "Test Email",
		Html:    "<p>Hello, this is a test email</p>",
	}

	expectedID := "email_12345"

	mockMailer.EXPECT().
		SendMail(gomock.Any(), gomock.Eq(emailConfig)).
		Return(expectedID, nil)

	id, err := mockMailer.SendMail(ctx, emailConfig)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)
}

func TestSendMailWithCompleteConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMailer := mocks.NewMockMailer(ctrl)

	ctx := context.Background()
	emailConfig := mailer.Config{
		From:    "test@example.com",
		To:      "recipient@example.com",
		Subject: "Test Email",
		Html:    "<p>Hello, this is a test email</p>",
		Cc:      []string{"cc@example.com"},
		Bcc:     []string{"bcc@example.com"},
		ReplyTo: "reply@example.com",
	}

	expectedID := "email_12345"

	mockMailer.EXPECT().
		SendMail(gomock.Any(), gomock.Eq(emailConfig)).
		Return(expectedID, nil)

	id, err := mockMailer.SendMail(ctx, emailConfig)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)
}

func TestSendMailFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMailer := mocks.NewMockMailer(ctrl)

	ctx := context.Background()
	emailConfig := mailer.Config{
		From:    "test@example.com",
		To:      "recipient@example.com",
		Subject: "Test Email",
		Html:    "<p>Hello, this is a test email</p>",
	}

	expectedErr := assert.AnError

	mockMailer.EXPECT().
		SendMail(gomock.Any(), gomock.Eq(emailConfig)).
		Return("", expectedErr)

	id, err := mockMailer.SendMail(ctx, emailConfig)

	assert.Error(t, err)
	assert.Equal(t, "", id)
	assert.Equal(t, expectedErr, err)
}
