package commands

type CreatePlanCommand struct {
	Name          string
	Description   string
	Price         float64
	Currency      string
	BillingPeriod string
	Features      []PlanFeatureInput
}

type UpdatePlanCommand struct {
	PlanID        string
	Name          string
	Description   string
	Price         float64
	Currency      string
	BillingPeriod string
}

type PlanFeatureInput struct {
	FeatureCode  string
	FeatureName  string
	FeatureValue string
}
