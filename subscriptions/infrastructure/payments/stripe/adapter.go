// Package stripe is the infrastructure adapter that implements the StripeService
// port using the real Stripe SDK. It is the only place that imports the Stripe
// library; everything else depends on the port interface instead.
package stripe

import (
	"errors"

	stripe "github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/subscription"
	"microservice-subscriptions-service/subscriptions/domain/model/entities"
)

// Adapter wraps the Stripe SDK. The "enabled" flag is a neat trick: when no
// secret key is configured (e.g. local development) the adapter becomes a no-op,
// so the service can run end to end without real Stripe credentials.
type Adapter struct{ enabled bool }

// NewAdapter sets the global Stripe key and enables the adapter only when a key
// is provided; otherwise it returns a disabled (no-op) adapter.
func NewAdapter(secretKey string) *Adapter {
	if secretKey != "" {
		stripe.Key = secretKey
		return &Adapter{enabled: true}
	}
	return &Adapter{enabled: false}
}

// CreateSubscription creates a subscription in Stripe and returns its id. The
// first Subscription argument is unused here (named "_"), but it is part of the
// port's signature so other implementations could use it.
func (a *Adapter) CreateSubscription(_ entities.Subscription, stripeCustomerID string, stripePriceID string) (string, error) {
	// Disabled adapter: do nothing and return an empty id (no error). The caller
	// treats an empty id as "not linked to Stripe".
	if !a.enabled {
		return "", nil
	}
	if stripeCustomerID == "" || stripePriceID == "" {
		return "", errors.New("stripe_customer_id and stripe_price_id are required for stripe subscription")
	}

	// A Stripe subscription needs a customer and at least one item (a price).
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

// CancelSubscription cancels the subscription in Stripe. It safely no-ops when
// the adapter is disabled or there is no Stripe id to cancel.
func (a *Adapter) CancelSubscription(stripeSubscriptionID string) error {
	if !a.enabled || stripeSubscriptionID == "" {
		return nil
	}
	_, err := subscription.Cancel(stripeSubscriptionID, nil)
	return err
}

// ChangePlan moves a Stripe subscription to a new price. Stripe does not let you
// just set a new price directly: you must update the EXISTING line item and
// point it at the new price. So the steps are:
//  1. Fetch the current subscription to find its first item's id.
//  2. Update that item to reference the target price.
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
	// Defensive check: a subscription should always have items, but if it does
	// not there is nothing to update.
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
