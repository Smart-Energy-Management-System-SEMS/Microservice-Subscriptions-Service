// Package outboundservices declares the "ports" the application uses to reach
// the outside world. In hexagonal (ports & adapters) architecture, a port is an
// interface owned by the inner layers; the concrete "adapter" lives in the
// infrastructure layer. StripeService is the port for the payment provider.
package outboundservices

import "microservice-subscriptions-service/subscriptions/domain/model/entities"

// StripeService is the contract the application depends on to manage
// subscriptions in Stripe. Because the application only knows this interface
// (not the Stripe SDK), the real adapter can be swapped for a fake in tests, and
// the rest of the code stays free of third-party imports.
//
// Note CreateSubscription returns a string: the Stripe subscription id, which we
// store locally to link our record to Stripe's.
type StripeService interface {
	CreateSubscription(subscription entities.Subscription, stripeCustomerID string, stripePriceID string) (string, error)
	CancelSubscription(stripeSubscriptionID string) error
	ChangePlan(stripeSubscriptionID, targetPriceID string) error
}
