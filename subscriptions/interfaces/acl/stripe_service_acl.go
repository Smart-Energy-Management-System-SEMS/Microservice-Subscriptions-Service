package acl

import (
	"errors"
	"strings"

	"microservice-subscriptions-service/subscriptions/application/outboundservices"
	"microservice-subscriptions-service/subscriptions/domain/model/entities"
)

// StripeServiceACL desacopla reglas internas del contrato externo de Stripe.
// Si no hay customer id, se omite la llamada externa y se permite la suscripcion local.
type StripeServiceACL struct {
	next outboundservices.StripeService
}

func NewStripeServiceACL(next outboundservices.StripeService) *StripeServiceACL {
	return &StripeServiceACL{next: next}
}

func (a *StripeServiceACL) CreateSubscription(subscription entities.Subscription, stripeCustomerID string, stripePriceID string) (string, error) {
	if a.next == nil {
		return "", nil
	}

	customerID := strings.TrimSpace(stripeCustomerID)
	if customerID == "" {
		return "", nil
	}

	priceID := strings.TrimSpace(stripePriceID)
	if priceID == "" {
		return "", errors.New("stripe price id is required when stripe_customer_id is provided")
	}

	return a.next.CreateSubscription(subscription, customerID, priceID)
}

func (a *StripeServiceACL) CancelSubscription(stripeSubscriptionID string) error {
	if a.next == nil {
		return nil
	}
	return a.next.CancelSubscription(strings.TrimSpace(stripeSubscriptionID))
}

func (a *StripeServiceACL) ChangePlan(stripeSubscriptionID, targetPriceID string) error {
	if a.next == nil {
		return nil
	}

	subscriptionID := strings.TrimSpace(stripeSubscriptionID)
	if subscriptionID == "" {
		return nil
	}

	priceID := strings.TrimSpace(targetPriceID)
	if priceID == "" {
		return errors.New("target stripe price id is required")
	}

	return a.next.ChangePlan(subscriptionID, priceID)
}
