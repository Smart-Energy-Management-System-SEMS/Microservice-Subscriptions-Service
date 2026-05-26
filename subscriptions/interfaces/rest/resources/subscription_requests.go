package resources

type CreateSubscriptionRequest struct {
	UserID string `json:"user_id" binding:"required"`
	PlanID string `json:"plan_id" binding:"required"`
}

type ChangePlanRequest struct {
	NewPlanID string `json:"new_plan_id" binding:"required"`
}
