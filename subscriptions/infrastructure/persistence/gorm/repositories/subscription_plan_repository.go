package repositories

import (
	"errors"

	"gorm.io/gorm"

	"microservice-subscriptions-service/subscriptions/domain/model/entities"
	domainrepo "microservice-subscriptions-service/subscriptions/domain/repositories"
	"microservice-subscriptions-service/subscriptions/infrastructure/persistence/gorm/model"
)

type SubscriptionPlanRepository struct{ db *gorm.DB }

func NewSubscriptionPlanRepository(db *gorm.DB) domainrepo.SubscriptionPlanRepository {
	return &SubscriptionPlanRepository{db: db}
}

func (r *SubscriptionPlanRepository) Create(plan *entities.SubscriptionPlan) error {
	m := toPlanModel(plan)
	return r.db.Create(m).Error
}

func (r *SubscriptionPlanRepository) Update(plan *entities.SubscriptionPlan) error {
	m := toPlanModel(plan)
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(m).Error
}

func (r *SubscriptionPlanRepository) Deactivate(planID string) error {
	res := r.db.Model(&model.SubscriptionPlanModel{}).Where("plan_id = ?", planID).Update("active", false)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *SubscriptionPlanRepository) FindAll() ([]entities.SubscriptionPlan, error) {
	var plans []model.SubscriptionPlanModel
	if err := r.db.Preload("Features").Find(&plans).Error; err != nil {
		return nil, err
	}
	out := make([]entities.SubscriptionPlan, 0, len(plans))
	for _, p := range plans {
		out = append(out, toPlanEntity(p))
	}
	return out, nil
}

func (r *SubscriptionPlanRepository) FindByID(planID string) (*entities.SubscriptionPlan, error) {
	var p model.SubscriptionPlanModel
	if err := r.db.Preload("Features").Where("plan_id = ?", planID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	e := toPlanEntity(p)
	return &e, nil
}

func toPlanModel(plan *entities.SubscriptionPlan) *model.SubscriptionPlanModel {
	m := &model.SubscriptionPlanModel{PlanID: plan.PlanID, Name: plan.Name, Description: plan.Description, Price: plan.Price, Currency: plan.Currency, BillingPeriod: plan.BillingPeriod, Active: plan.Active, CreatedAt: plan.CreatedAt}
	features := make([]model.PlanFeatureModel, 0, len(plan.PlanFeatures))
	for _, f := range plan.PlanFeatures {
		features = append(features, model.PlanFeatureModel{FeatureID: f.FeatureID, PlanID: f.PlanID, FeatureCode: f.FeatureCode, FeatureName: f.FeatureName, FeatureValue: f.FeatureValue, CreatedAt: f.CreatedAt})
	}
	m.Features = features
	return m
}

func toPlanEntity(m model.SubscriptionPlanModel) entities.SubscriptionPlan {
	features := make([]entities.PlanFeature, 0, len(m.Features))
	for _, f := range m.Features {
		features = append(features, entities.PlanFeature{FeatureID: f.FeatureID, PlanID: f.PlanID, FeatureCode: f.FeatureCode, FeatureName: f.FeatureName, FeatureValue: f.FeatureValue, CreatedAt: f.CreatedAt})
	}
	return entities.SubscriptionPlan{PlanID: m.PlanID, Name: m.Name, Description: m.Description, Price: m.Price, Currency: m.Currency, BillingPeriod: m.BillingPeriod, Active: m.Active, CreatedAt: m.CreatedAt, PlanFeatures: features}
}
