package outboundservices

import "microservice-subscriptions-service/subscriptions/domain/model/entities"

type StripeService interface {
	CreateSubscription(subscription entities.Subscription) (string, error)
	CancelSubscription(stripeSubscriptionID string) error
	ChangePlan(stripeSubscriptionID, targetPlanExternalCode string) error
}
