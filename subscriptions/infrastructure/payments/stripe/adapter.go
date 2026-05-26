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

func (a *Adapter) CreateSubscription(_ entities.Subscription, stripeCustomerID string, stripePriceID string) (string, error) {
	if !a.enabled {
		return "", nil
	}
	if stripeCustomerID == "" || stripePriceID == "" {
		return "", errors.New("stripe_customer_id and stripe_price_id are required for stripe subscription")
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(stripeCustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{Price: stripe.String(stripePriceID)},
		},
	}
	sub, err := subscription.New(params)
	if err != nil {
		return "", err
	}
	return sub.ID, nil
}

func (a *Adapter) CancelSubscription(stripeSubscriptionID string) error {
	if !a.enabled || stripeSubscriptionID == "" {
		return nil
	}
	_, err := subscription.Cancel(stripeSubscriptionID, nil)
	return err
}

func (a *Adapter) ChangePlan(stripeSubscriptionID, targetPriceID string) error {
	if !a.enabled || stripeSubscriptionID == "" {
		return nil
	}
	if targetPriceID == "" {
		return errors.New("target stripe price id is required")
	}
	current, err := subscription.Get(stripeSubscriptionID, nil)
	if err != nil {
		return err
	}
	if len(current.Items.Data) == 0 {
		return errors.New("stripe subscription has no items to update")
	}
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(current.Items.Data[0].ID),
				Price: stripe.String(targetPriceID),
			},
		},
	}
	_, err = subscription.Update(stripeSubscriptionID, params)
	return err
}
