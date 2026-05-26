package model

import "time"

type SubscriptionPlanModel struct {
	PlanID        string             `gorm:"type:uuid;primaryKey;column:plan_id"`
	Name          string             `gorm:"size:100;not null"`
	Description   string             `gorm:"type:text"`
	Price         float64            `gorm:"type:numeric(10,2);not null"`
	Currency      string             `gorm:"size:10;not null"`
	BillingPeriod string             `gorm:"size:30;not null;column:billing_period"`
	Active        bool               `gorm:"default:true"`
	CreatedAt     time.Time          `gorm:"column:created_at"`
	Features      []PlanFeatureModel `gorm:"foreignKey:PlanID;references:PlanID"`
}

func (SubscriptionPlanModel) TableName() string { return "subscription_plans" }
