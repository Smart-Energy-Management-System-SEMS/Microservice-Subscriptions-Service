package integrationevents

import (
	"encoding/json"
	"time"

	"microservice-subscriptions-service/subscriptions/domain/model/entities"
)

const (
	EventTypeSubscriptionCreated          = "subscription.created"
	EventTypeSubscriptionUpdated          = "subscription.updated"
	EventTypeSubscriptionCancelled        = "subscription.cancelled"
	EventTypeSubscriptionExpired          = "subscription.expired"
	EventTypeSubscriptionPlanChanged      = "subscription.plan.changed"
	EventTypeSubscriptionRenewalRequested = "subscription.renewal.requested"
	DefaultSubscriptionsTopic             = "subscriptions.events"
)

type SubscriptionEvent struct {
	EventType  string         `json:"eventType"`
	UserID     string         `json:"userId"`
	OccurredAt time.Time      `json:"occurredAt"`
	Data       map[string]any `json:"data"`
}

func MarshalSubscriptionEvent(eventType string, occurredAt time.Time, subscription *entities.Subscription, extraData map[string]any) ([]byte, error) {
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	data := map[string]any{
		"subscriptionId": subscription.SubscriptionID,
		"userId":         subscription.UserID,
		"planId":         subscription.PlanID,
		"status":         subscription.Status,
		"startDate":      subscription.StartDate,
		"endDate":        subscription.EndDate,
		"createdAt":      subscription.CreatedAt,
	}
	if subscription.StripeSubscriptionID != nil {
		data["stripeSubscriptionId"] = *subscription.StripeSubscriptionID
	}
	for key, value := range extraData {
		data[key] = value
	}

	return json.Marshal(SubscriptionEvent{
		EventType:  eventType,
		UserID:     subscription.UserID,
		OccurredAt: occurredAt.UTC(),
		Data:       data,
	})
}
