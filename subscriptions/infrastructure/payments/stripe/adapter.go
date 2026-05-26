package stripe

import (
	"errors"

	stripe "github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/subscription"
	"microservice-subscriptions-service/subscriptions/domain/model/entities"
)

type Adapter struct{ enabled bool }

func NewAdapter(secretKey string) *Adapter {
	if secretKey != "" {
		stripe.Key = secretKey
		return &Adapter{enabled: true}
	}
	return &Adapter{enabled: false}
}

func (a *Adapter) CreateSubscription(_ entities.Subscription) (string, error) {
	if !a.enabled {
		return "", nil
	}
	return "", errors.New("stripe create subscription requires customer and price mapping")
}

func (a *Adapter) CancelSubscription(stripeSubscriptionID string) error {
	if !a.enabled || stripeSubscriptionID == "" {
		return nil
	}
	_, err := subscription.Cancel(stripeSubscriptionID, nil)
	return err
}

func (a *Adapter) ChangePlan(stripeSubscriptionID, _ string) error {
	if !a.enabled || stripeSubscriptionID == "" {
		return nil
	}
	return errors.New("stripe change plan requires item id and price mapping")
}
