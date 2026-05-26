package model

import "time"

type PlanFeatureModel struct {
	FeatureID    string    `gorm:"type:uuid;primaryKey;column:feature_id"`
	PlanID       string    `gorm:"type:uuid;index;column:plan_id"`
	FeatureCode  string    `gorm:"size:60;not null;column:feature_code"`
	FeatureName  string    `gorm:"size:120;not null;column:feature_name"`
	FeatureValue string    `gorm:"size:100;not null;column:feature_value"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (PlanFeatureModel) TableName() string { return "plan_features" }
