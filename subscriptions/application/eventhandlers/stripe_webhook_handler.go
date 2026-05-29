package eventhandlers

import (
	"encoding/json"
	"strings"
	"time"

	stripe "github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/webhook"
	"microservice-subscriptions-service/subscriptions/application/outboundservices"
	"microservice-subscriptions-service/subscriptions/domain/model/valueobjects"
	domainrepo "microservice-subscriptions-service/subscriptions/domain/repositories"
)

type StripeWebhookHandler struct {
	subscriptions domainrepo.SubscriptionRepository
	events        outboundservices.EventPublisher
	topics        StripeWebhookTopics
}

type StripeWebhookTopics struct {
	Expired string
	Updated string
}

func NewStripeWebhookHandler(subscriptions domainrepo.SubscriptionRepository, events outboundservices.EventPublisher, topics StripeWebhookTopics) *StripeWebhookHandler {
	if strings.TrimSpace(topics.Expired) == "" {
		topics.Expired = "SubscriptionExpired"
	}
	if strings.TrimSpace(topics.Updated) == "" {
		topics.Updated = "SubscriptionUpdated"
	}
	return &StripeWebhookHandler{subscriptions: subscriptions, events: events, topics: topics}
}

func (h *StripeWebhookHandler) Handle(payload []byte, signature string, secret string) error {
	event, err := webhook.ConstructEvent(payload, signature, secret)
	if err != nil {
		return err
	}

	switch event.Type {
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		return h.handleSubscriptionEvent(event)
	default:
		return nil
	}
}

func (h *StripeWebhookHandler) handleSubscriptionEvent(event stripe.Event) error {
	var stripeSubscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSubscription); err != nil {
		return err
	}
	if stripeSubscription.ID == "" {
		return nil
	}

	subscription, err := h.subscriptions.FindByStripeSubscriptionID(stripeSubscription.ID)
	if err != nil || subscription == nil {
		return err
	}

	subscription.Status = mapStripeStatus(stripeSubscription.Status)
	if stripeSubscription.CancelAtPeriodEnd || subscription.Status == valueobjects.StatusCancelled || subscription.Status == valueobjects.StatusExpired {
		end := time.Unix(stripeSubscription.CurrentPeriodEnd, 0).UTC()
		subscription.EndDate = &end
	}
	if err = h.subscriptions.Update(subscription); err != nil {
		return err
	}

	if h.events != nil {
		payload, marshalErr := json.Marshal(subscription)
		if marshalErr == nil {
			switch event.Type {
			case "customer.subscription.deleted":
				_ = h.events.Publish(h.topics.Expired, payload)
			case "customer.subscription.updated":
				_ = h.events.Publish(h.topics.Updated, payload)
			}
		}
	}

	return nil
}

func mapStripeStatus(status stripe.SubscriptionStatus) valueobjects.SubscriptionStatus {
	switch status {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing:
		return valueobjects.StatusActive
	case stripe.SubscriptionStatusCanceled:
		return valueobjects.StatusCancelled
	case stripe.SubscriptionStatusIncompleteExpired:
		return valueobjects.StatusExpired
	case stripe.SubscriptionStatusPastDue, stripe.SubscriptionStatusUnpaid, stripe.SubscriptionStatusIncomplete:
		return valueobjects.StatusPendingRenewal
	default:
		return valueobjects.StatusInactive
	}
}
