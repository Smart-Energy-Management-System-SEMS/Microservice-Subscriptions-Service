package queryservices

import (
	"microservice-subscriptions-service/subscriptions/domain/model/entities"
	domainrepo "microservice-subscriptions-service/subscriptions/domain/repositories"
)

type PlanQueryService struct {
	repo domainrepo.SubscriptionPlanRepository
}

func NewPlanQueryService(repo domainrepo.SubscriptionPlanRepository) *PlanQueryService {
	return &PlanQueryService{repo: repo}
}

func (s *PlanQueryService) FindAll() ([]entities.SubscriptionPlan, error) { return s.repo.FindAll() }
func (s *PlanQueryService) FindByID(planID string) (*entities.SubscriptionPlan, error) {
	return s.repo.FindByID(planID)
}
