package repositories

import (
	"errors"

	"gorm.io/gorm"

	"microservice-subscriptions-service/subscriptions/domain/model/entities"
	"microservice-subscriptions-service/subscriptions/domain/model/valueobjects"
	domainrepo "microservice-subscriptions-service/subscriptions/domain/repositories"
	"microservice-subscriptions-service/subscriptions/infrastructure/persistence/gorm/model"
)

// SubscriptionRepository is the GORM (database) implementation of the domain's
// SubscriptionRepository interface. It is an adapter: the domain declares WHAT
// persistence operations exist, this struct implements HOW they run against SQL.
type SubscriptionRepository struct{ db *gorm.DB }

// NewSubscriptionRepository returns the domain interface type, not the concrete
// struct. Returning the interface keeps the dependency pointing the right way
// (infrastructure depends on the domain, never the reverse).
func NewSubscriptionRepository(db *gorm.DB) domainrepo.SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create inserts a new subscription. toSubscriptionModel converts the domain
// entity into the DB-shaped struct first.
func (r *SubscriptionRepository) Create(subscription *entities.Subscription) error {
	m := toSubscriptionModel(subscription)
	return r.db.Create(m).Error
}

// Update writes changes to an existing row. We inspect RowsAffected: if the
// UPDATE matched no rows, the subscription does not exist, so we return a clear
// "not found" error instead of silently succeeding.
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

// FindByID loads one subscription. Design choice worth noting: a "not found" is
// reported as (nil, nil) — no entity, no error — so the caller decides how to
// treat a missing record. Any OTHER database error is returned as a real error.
// The "?" placeholder is a parameterised query, which prevents SQL injection.
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

func (r *SubscriptionRepository) FindByStripeSubscriptionID(stripeSubscriptionID string) (*entities.Subscription, error) {
	var m model.SubscriptionModel
	if err := r.db.Where("stripe_subscription_id = ?", stripeSubscriptionID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	e := toSubscriptionEntity(m)
	return &e, nil
}

// toSubscriptionModel and toSubscriptionEntity are the "mappers" that translate
// between the domain entity and the database model. Keeping the two shapes
// separate lets the DB schema and the domain evolve independently. The status
// is stored as a plain string column, so we cast it both ways.

func toSubscriptionModel(e *entities.Subscription) *model.SubscriptionModel {
	return &model.SubscriptionModel{SubscriptionID: e.SubscriptionID, UserID: e.UserID, PlanID: e.PlanID, Status: string(e.Status), StartDate: e.StartDate, EndDate: e.EndDate, StripeSubscriptionID: e.StripeSubscriptionID, CreatedAt: e.CreatedAt}
}

// toSubscriptionEntity rebuilds the domain entity from a row. It also defends
// against bad data: if the stored status is not one we recognise, we fall back
// to INACTIVE instead of letting an invalid value propagate through the system.
func toSubscriptionEntity(m model.SubscriptionModel) entities.Subscription {
	status := valueobjects.SubscriptionStatus(m.Status)
	if !valueobjects.IsValidStatus(status) {
		status = valueobjects.StatusInactive
	}
	return entities.Subscription{SubscriptionID: m.SubscriptionID, UserID: m.UserID, PlanID: m.PlanID, Status: status, StartDate: m.StartDate, EndDate: m.EndDate, StripeSubscriptionID: m.StripeSubscriptionID, CreatedAt: m.CreatedAt}
}
