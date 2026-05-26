package model

import "time"

type SubscriptionModel struct {
	SubscriptionID       string     `gorm:"type:uuid;primaryKey;column:subscription_id"`
	UserID               string     `gorm:"type:uuid;index;not null;column:user_id"`
	PlanID               string     `gorm:"type:uuid;not null;column:plan_id"`
	Status               string     `gorm:"size:30;not null"`
	StartDate            time.Time  `gorm:"type:date;not null;column:start_date"`
	EndDate              *time.Time `gorm:"type:date;column:end_date"`
	StripeSubscriptionID *string    `gorm:"size:150;column:stripe_subscription_id"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
}

func (SubscriptionModel) TableName() string { return "subscriptions" }
