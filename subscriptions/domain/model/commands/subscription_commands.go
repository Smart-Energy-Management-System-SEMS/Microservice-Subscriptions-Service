package commands

type CreateSubscriptionCommand struct {
	UserID string
	PlanID string
}

type CancelSubscriptionCommand struct {
	SubscriptionID string
}

type ChangePlanCommand struct {
	SubscriptionID string
	NewPlanID      string
}
