package repositories

import "microservice-subscriptions-service/subscriptions/domain/model/entities"

type SubscriptionRepository interface {
	Create(subscription *entities.Subscription) error
	Update(subscription *entities.Subscription) error
	FindByID(subscriptionID string) (*entities.Subscription, error)
	FindByUserID(userID string) ([]entities.Subscription, error)
	FindByStripeSubscriptionID(stripeSubscriptionID string) (*entities.Subscription, error)
}
