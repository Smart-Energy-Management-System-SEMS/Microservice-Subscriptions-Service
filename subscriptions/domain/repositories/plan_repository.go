package repositories

import "microservice-subscriptions-service/subscriptions/domain/model/entities"

type SubscriptionPlanRepository interface {
	Create(plan *entities.SubscriptionPlan) error
	Update(plan *entities.SubscriptionPlan) error
	Deactivate(planID string) error
	FindAll() ([]entities.SubscriptionPlan, error)
	FindByID(planID string) (*entities.SubscriptionPlan, error)
}
