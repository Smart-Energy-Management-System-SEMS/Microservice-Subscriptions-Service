package resources

type CreatePlanRequest struct {
	Name          string               `json:"name" binding:"required"`
	Description   string               `json:"description"`
	Price         float64              `json:"price" binding:"required"`
	Currency      string               `json:"currency" binding:"required"`
	BillingPeriod string               `json:"billing_period" binding:"required"`
	Features      []PlanFeatureRequest `json:"features"`
}

type UpdatePlanRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description"`
	Price         float64 `json:"price" binding:"required"`
	Currency      string  `json:"currency" binding:"required"`
	BillingPeriod string  `json:"billing_period" binding:"required"`
}

type PlanFeatureRequest struct {
	FeatureCode  string `json:"feature_code"`
	FeatureName  string `json:"feature_name"`
	FeatureValue string `json:"feature_value"`
}
