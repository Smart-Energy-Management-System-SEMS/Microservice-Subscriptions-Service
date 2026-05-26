package commands

type CreateSubscriptionCommand struct {
	UserID           string
	PlanID           string
	StripeCustomerID string
}

type CancelSubscriptionCommand struct {
	SubscriptionID string
}

type ChangePlanCommand struct {
	SubscriptionID string
	NewPlanID      string
}
