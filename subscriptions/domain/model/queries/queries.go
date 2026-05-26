package queries

type GetPlanByIDQuery struct {
	PlanID string
}

type GetSubscriptionByIDQuery struct {
	SubscriptionID string
}

type GetSubscriptionsByUserIDQuery struct {
	UserID string
}
