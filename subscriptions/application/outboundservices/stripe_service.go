package outboundservices

import "microservice-subscriptions-service/subscriptions/domain/model/entities"

type StripeService interface {
	CreateSubscription(subscription entities.Subscription, stripeCustomerID string, stripePriceID string) (string, error)
	CancelSubscription(stripeSubscriptionID string) error
	ChangePlan(stripeSubscriptionID, targetPriceID string) error
}
