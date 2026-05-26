package services

import (
	"errors"

	"microservice-subscriptions-service/subscriptions/domain/model/valueobjects"
)

type SubscriptionManager struct{}

func NewSubscriptionManager() *SubscriptionManager { return &SubscriptionManager{} }

func (m *SubscriptionManager) CanCancel(status valueobjects.SubscriptionStatus) error {
	if status == valueobjects.StatusCancelled || status == valueobjects.StatusExpired {
		return errors.New("subscription cannot be cancelled from current status")
	}
	return nil
}

func (m *SubscriptionManager) CanChangePlan(status valueobjects.SubscriptionStatus) error {
	if status == valueobjects.StatusCancelled || status == valueobjects.StatusExpired {
		return errors.New("subscription cannot change plan from current status")
	}
	return nil
}
