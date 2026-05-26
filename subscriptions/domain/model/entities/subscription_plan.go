package entities

import "time"

type SubscriptionPlan struct {
	PlanID        string
	Name          string
	Description   string
	Price         float64
	Currency      string
	BillingPeriod string
	Active        bool
	CreatedAt     time.Time
	PlanFeatures  []PlanFeature
}

type PlanFeature struct {
	FeatureID    string
	PlanID       string
	FeatureCode  string
	FeatureName  string
	FeatureValue string
	CreatedAt    time.Time
}
