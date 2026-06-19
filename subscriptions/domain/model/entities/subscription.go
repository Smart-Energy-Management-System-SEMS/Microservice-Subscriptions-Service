// Package entities holds the core domain entities. An entity is an object with
// its own identity (here, the SubscriptionID): two subscriptions are "the same"
// when their ids match, even if their other fields differ.
package entities

import (
	"time"

	"microservice-subscriptions-service/subscriptions/domain/model/valueobjects"
)

// Subscription represents a user's subscription to a plan. It is a plain domain
// struct with no database or framework annotations, which keeps the domain
// independent of infrastructure and easy to unit test.
//
// Two fields are pointers because they are OPTIONAL, and Go uses a nil pointer
// to mean "no value":
//   - EndDate is nil while the subscription is still ongoing (it only gets a
//     value once cancelled/expired).
//   - StripeSubscriptionID is nil when the subscription is not linked to Stripe
//     (for example, when Stripe is disabled in local development).
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
