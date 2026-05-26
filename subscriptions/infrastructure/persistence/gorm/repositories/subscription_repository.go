package repositories

import (
	"errors"

	"gorm.io/gorm"

	"microservice-subscriptions-service/subscriptions/domain/model/entities"
	"microservice-subscriptions-service/subscriptions/domain/model/valueobjects"
	domainrepo "microservice-subscriptions-service/subscriptions/domain/repositories"
	"microservice-subscriptions-service/subscriptions/infrastructure/persistence/gorm/model"
)

type SubscriptionRepository struct{ db *gorm.DB }

func NewSubscriptionRepository(db *gorm.DB) domainrepo.SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(subscription *entities.Subscription) error {
	m := toSubscriptionModel(subscription)
	return r.db.Create(m).Error
}

func (r *SubscriptionRepository) Update(subscription *entities.Subscription) error {
	m := toSubscriptionModel(subscription)
	res := r.db.Model(&model.SubscriptionModel{}).Where("subscription_id = ?", m.SubscriptionID).Updates(m)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *SubscriptionRepository) FindByID(subscriptionID string) (*entities.Subscription, error) {
	var m model.SubscriptionModel
	if err := r.db.Where("subscription_id = ?", subscriptionID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	e := toSubscriptionEntity(m)
	return &e, nil
}

func (r *SubscriptionRepository) FindByUserID(userID string) ([]entities.Subscription, error) {
	var rows []model.SubscriptionModel
	if err := r.db.Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entities.Subscription, 0, len(rows))
	for _, m := range rows {
		out = append(out, toSubscriptionEntity(m))
	}
	return out, nil
}

func toSubscriptionModel(e *entities.Subscription) *model.SubscriptionModel {
	return &model.SubscriptionModel{SubscriptionID: e.SubscriptionID, UserID: e.UserID, PlanID: e.PlanID, Status: string(e.Status), StartDate: e.StartDate, EndDate: e.EndDate, StripeSubscriptionID: e.StripeSubscriptionID, CreatedAt: e.CreatedAt}
}

func toSubscriptionEntity(m model.SubscriptionModel) entities.Subscription {
	status := valueobjects.SubscriptionStatus(m.Status)
	if !valueobjects.IsValidStatus(status) {
		status = valueobjects.StatusInactive
	}
	return entities.Subscription{SubscriptionID: m.SubscriptionID, UserID: m.UserID, PlanID: m.PlanID, Status: status, StartDate: m.StartDate, EndDate: m.EndDate, StripeSubscriptionID: m.StripeSubscriptionID, CreatedAt: m.CreatedAt}
}
