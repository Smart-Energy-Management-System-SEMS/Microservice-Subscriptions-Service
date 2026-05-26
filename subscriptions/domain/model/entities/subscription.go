package entities

import (
	"time"

	"microservice-subscriptions-service/subscriptions/domain/model/valueobjects"
)

type Subscription struct {
	SubscriptionID       string
	UserID               string
	PlanID               string
	Status               valueobjects.SubscriptionStatus
	StartDate            time.Time
	EndDate              *time.Time
	StripeSubscriptionID *string
	CreatedAt            time.Time
}
