package transform

import (
	"microservice-subscriptions-service/subscriptions/domain/model/commands"
	"microservice-subscriptions-service/subscriptions/interfaces/rest/resources"
)

func ToCreatePlanCommand(r resources.CreatePlanRequest) commands.CreatePlanCommand {
	features := make([]commands.PlanFeatureInput, 0, len(r.Features))
	for _, f := range r.Features {
		features = append(features, commands.PlanFeatureInput{FeatureCode: f.FeatureCode, FeatureName: f.FeatureName, FeatureValue: f.FeatureValue})
	}
	return commands.CreatePlanCommand{Name: r.Name, Description: r.Description, Price: r.Price, Currency: r.Currency, BillingPeriod: r.BillingPeriod, Features: features}
}

func ToUpdatePlanCommand(planID string, r resources.UpdatePlanRequest) commands.UpdatePlanCommand {
	return commands.UpdatePlanCommand{PlanID: planID, Name: r.Name, Description: r.Description, Price: r.Price, Currency: r.Currency, BillingPeriod: r.BillingPeriod}
}

func ToCreateSubscriptionCommand(r resources.CreateSubscriptionRequest) commands.CreateSubscriptionCommand {
	return commands.CreateSubscriptionCommand{UserID: r.UserID, PlanID: r.PlanID, StripeCustomerID: r.StripeCustomerID}
}
