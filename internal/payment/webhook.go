package payment

import (
	"context"
	"encoding/json"
	"time"

	"github.com/stripe/stripe-go/v81"
)

func (s *Service) handleInvoicePaymentSucceeded(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return ErrInvalidInvoicePayload
	}

	if len(invoice.Lines.Data) == 0 || invoice.Lines.Data[0].Price == nil {
		return ErrInvalidSubscriptionData
	}

	priceID := invoice.Lines.Data[0].Price.ID
	subscriptionType, ok := planMap[priceID]
	if !ok {
		return ErrUnknownPriceID
	}

	sub, err := s.client.Subscriptions.Get(invoice.Subscription.ID, nil)
	if err != nil {
		return ErrStripeSubscriptionFetch
	}

	return s.updateSubscription(ctx, sub, priceID, subscriptionType, string(sub.Status))
}

func (s *Service) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return ErrInvalidSubscriptionPayload
	}

	return s.handleSubscriptionEvent(ctx, &sub, string(sub.Status))
}

func (s *Service) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return ErrInvalidSubscriptionPayload
	}

	return s.handleSubscriptionEvent(ctx, &sub, string(sub.Status))
}

func (s *Service) handleCheckoutSessionCompleted(ctx context.Context, event stripe.Event) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return ErrInvalidSubscriptionPayload
	}

	if session.Customer == nil || session.Subscription == nil {
		return ErrCheckoutSessionInvalid
	}

	sub, err := s.client.Subscriptions.Get(session.Subscription.ID, nil)
	if err != nil {
		return ErrStripeSubscriptionFetch
	}

	return s.handleSubscriptionEvent(ctx, sub, string(sub.Status))
}

func (s *Service) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return ErrInvalidInvoicePayload
	}

	sub, err := s.client.Subscriptions.Get(invoice.Subscription.ID, nil)
	if err != nil {
		return ErrStripeSubscriptionFetch
	}

	return s.handleSubscriptionEvent(ctx, sub, "past_due")
}

func (s *Service) handleSubscriptionEvent(ctx context.Context, sub *stripe.Subscription, status string) error {
	if len(sub.Items.Data) == 0 || sub.Items.Data[0].Price == nil {
		return ErrInvalidSubscriptionData
	}

	priceID := sub.Items.Data[0].Price.ID
	subscriptionType, ok := planMap[priceID]
	if !ok {
		return ErrUnknownPriceID
	}

	return s.updateSubscription(ctx, sub, priceID, subscriptionType, status)
}

func (s *Service) updateSubscription(ctx context.Context, sub *stripe.Subscription, priceID string, subscriptionType SubscriptionType, status string) error {
	return s.r.UpdateUserSubscriptionFromStripeWebhook(
		ctx,
		sub.Customer.ID,
		subscriptionType,
		sub.ID,
		priceID,
		status,
		time.Unix(sub.CurrentPeriodStart, 0).UTC(),
		time.Unix(sub.CurrentPeriodEnd, 0).UTC(),
		sub.CancelAtPeriodEnd,
		toNullableTime(sub.CanceledAt),
		toNullableTime(sub.EndedAt),
	)
}

func toNullableTime(ts int64) *time.Time {
	if ts == 0 {
		return nil
	}
	t := time.Unix(ts, 0).UTC()
	return &t
}
