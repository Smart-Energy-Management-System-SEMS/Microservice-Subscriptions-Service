package commandservices

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"microservice-subscriptions-service/subscriptions/domain/model/commands"
	"microservice-subscriptions-service/subscriptions/domain/model/entities"
	domainrepo "microservice-subscriptions-service/subscriptions/domain/repositories"
)

type PlanCommandService struct {
	repo domainrepo.SubscriptionPlanRepository
}

func NewPlanCommandService(repo domainrepo.SubscriptionPlanRepository) *PlanCommandService {
	return &PlanCommandService{repo: repo}
}

func (s *PlanCommandService) Create(cmd commands.CreatePlanCommand) (*entities.SubscriptionPlan, error) {
	if cmd.Name == "" || cmd.Currency == "" || cmd.BillingPeriod == "" {
		return nil, errors.New("name, currency and billing_period are required")
	}
	planID := uuid.NewString()
	features := make([]entities.PlanFeature, 0, len(cmd.Features))
	for _, f := range cmd.Features {
		features = append(features, entities.PlanFeature{FeatureID: uuid.NewString(), PlanID: planID, FeatureCode: f.FeatureCode, FeatureName: f.FeatureName, FeatureValue: f.FeatureValue, CreatedAt: time.Now().UTC()})
	}
	plan := &entities.SubscriptionPlan{PlanID: planID, Name: cmd.Name, Description: cmd.Description, Price: cmd.Price, Currency: cmd.Currency, BillingPeriod: cmd.BillingPeriod, Active: true, CreatedAt: time.Now().UTC(), PlanFeatures: features}
	if err := s.repo.Create(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *PlanCommandService) Update(cmd commands.UpdatePlanCommand) (*entities.SubscriptionPlan, error) {
	plan, err := s.repo.FindByID(cmd.PlanID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, gorm.ErrRecordNotFound
	}
	plan.Name = cmd.Name
	plan.Description = cmd.Description
	plan.Price = cmd.Price
	plan.Currency = cmd.Currency
	plan.BillingPeriod = cmd.BillingPeriod
	if err = s.repo.Update(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *PlanCommandService) Deactivate(planID string) error { return s.repo.Deactivate(planID) }
