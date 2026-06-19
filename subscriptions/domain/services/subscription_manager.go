// Package services holds "domain services": business logic that does not
// naturally belong to a single entity. Here the manager centralises the rules
// about which state transitions are allowed, so those rules live in the domain
// (not scattered across the application layer).
package services

import (
	"errors"

	"microservice-subscriptions-service/subscriptions/domain/model/valueobjects"
)

// SubscriptionManager is an empty struct: it holds no data and only groups
// related domain rules. An empty struct{} uses zero memory, so this is a cheap,
// stateless service.
type SubscriptionManager struct{}

// NewSubscriptionManager is the constructor. We keep one for consistency even
// though the struct is empty, so dependencies could be added later without
// changing how callers create it.
func NewSubscriptionManager() *SubscriptionManager { return &SubscriptionManager{} }

// CanCancel encodes a guard rule: a subscription that is already cancelled or
// expired cannot be cancelled again. Returning an error (instead of a bool) lets
// the caller forward a clear, ready-to-show message.
func (m *SubscriptionManager) CanCancel(status valueobjects.SubscriptionStatus) error {
	if status == valueobjects.StatusCancelled || status == valueobjects.StatusExpired {
		return errors.New("subscription cannot be cancelled from current status")
	}
	return nil
}

// CanChangePlan applies the same idea to changing plans: a finished
// subscription (cancelled/expired) can no longer switch plans.
func (m *SubscriptionManager) CanChangePlan(status valueobjects.SubscriptionStatus) error {
	if status == valueobjects.StatusCancelled || status == valueobjects.StatusExpired {
		return errors.New("subscription cannot change plan from current status")
	}
	return nil
}
