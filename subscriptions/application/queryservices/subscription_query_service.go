package queryservices

import (
	"microservice-subscriptions-service/subscriptions/domain/model/entities"
	domainrepo "microservice-subscriptions-service/subscriptions/domain/repositories"
)

type SubscriptionQueryService struct {
	repo domainrepo.SubscriptionRepository
}

func NewSubscriptionQueryService(repo domainrepo.SubscriptionRepository) *SubscriptionQueryService {
	return &SubscriptionQueryService{repo: repo}
}

func (s *SubscriptionQueryService) FindByID(subscriptionID string) (*entities.Subscription, error) {
	return s.repo.FindByID(subscriptionID)
}
func (s *SubscriptionQueryService) FindByUserID(userID string) ([]entities.Subscription, error) {
	return s.repo.FindByUserID(userID)
}
