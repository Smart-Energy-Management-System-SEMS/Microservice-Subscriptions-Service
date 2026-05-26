package configuration

import (
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	model2 "microservice-subscriptions-service/subscriptions/infrastructure/persistence/gorm/model"
)

func NewDatabase(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model2.SubscriptionPlanModel{},
		&model2.PlanFeatureModel{},
		&model2.SubscriptionModel{},
	)
}
