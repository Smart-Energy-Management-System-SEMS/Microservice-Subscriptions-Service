package commandservices

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"microservice-subscriptions-service/subscriptions/application/integrationevents"
	"microservice-subscriptions-service/subscriptions/application/outboundservices"
	"microservice-subscriptions-service/subscriptions/domain/model/commands"
	"microservice-subscriptions-service/subscriptions/domain/model/entities"
	"microservice-subscriptions-service/subscriptions/domain/model/valueobjects"
	domainrepo "microservice-subscriptions-service/subscriptions/domain/repositories"
	"microservice-subscriptions-service/subscriptions/domain/services"
)

// SubscriptionCommandService holds the write use cases for subscriptions (the
// "command" side of CQRS, where operations change state). It lives in the
// application layer and orchestrates the work: it calls the repositories, the
// Stripe port and the event publisher, and asks the domain manager to validate
// state transitions. The business rules themselves stay in the domain.
//
// All collaborators are interfaces (repositories, StripeService, EventPublisher)
// — dependency injection — so real implementations can be swapped for fakes in
// tests.
type SubscriptionCommandService struct {
	subscriptions domainrepo.SubscriptionRepository
	plans         domainrepo.SubscriptionPlanRepository
	stripe        outboundservices.StripeService
	events        outboundservices.EventPublisher
	manager       *services.SubscriptionManager
	topics        SubscriptionTopics
}

type SubscriptionTopics struct {
	Events string
}

// NewSubscriptionCommandService wires the dependencies. The blocks below apply
// safe default topic names when none are configured, so the service still works
// out of the box.
func NewSubscriptionCommandService(subscriptions domainrepo.SubscriptionRepository, plans domainrepo.SubscriptionPlanRepository, stripe outboundservices.StripeService, events outboundservices.EventPublisher, topics SubscriptionTopics) *SubscriptionCommandService {
	if strings.TrimSpace(topics.Events) == "" {
		topics.Events = integrationevents.DefaultSubscriptionsTopic
	}
	return &SubscriptionCommandService{subscriptions: subscriptions, plans: plans, stripe: stripe, events: events, manager: services.NewSubscriptionManager(), topics: topics}
}

// Create is the use case for starting a new subscription. It reads top to bottom
// like a recipe and returns early on any error (Go's idiomatic style).
func (s *SubscriptionCommandService) Create(cmd commands.CreateSubscriptionCommand) (*entities.Subscription, error) {
	// Basic input validation first.
	if cmd.UserID == "" || cmd.PlanID == "" {
		return nil, errors.New("user_id and plan_id are required")
	}
	// The plan must exist and be active before anyone can subscribe to it.
	plan, err := s.plans.FindByID(cmd.PlanID)
	if err != nil {
		return nil, err
	}
	if plan == nil || !plan.Active {
		return nil, errors.New("plan is not available")
	}
	// Each plan stores its Stripe price id among its features; pull it out.
	stripePriceID, err := extractStripePriceID(*plan)
	if err != nil {
		return nil, err
	}
	// Build the new subscription. It starts ACTIVE with a freshly generated id
	// and UTC timestamps (UTC keeps times consistent across servers/timezones).
	subscription := &entities.Subscription{SubscriptionID: uuid.NewString(), UserID: cmd.UserID, PlanID: cmd.PlanID, Status: valueobjects.StatusActive, StartDate: time.Now().UTC(), CreatedAt: time.Now().UTC()}
	// Try to create the matching subscription in Stripe. Note this is tolerant:
	// we only attach the Stripe id when the call succeeds AND returns one. If
	// Stripe is disabled or fails, we still create the local subscription rather
	// than blocking the user.
	stripeID, err := s.stripe.CreateSubscription(*subscription, cmd.StripeCustomerID, stripePriceID)
	if err == nil && stripeID != "" {
		subscription.StripeSubscriptionID = &stripeID
	}
	// Persist, then announce the event so other services can react.
	if err = s.subscriptions.Create(subscription); err != nil {
		return nil, err
	}
	s.publish(integrationevents.EventTypeSubscriptionCreated, subscription, plan, nil)
	return subscription, nil
}

// Cancel follows the classic "load, validate, change, save" pattern. The domain
// manager decides whether cancelling is allowed from the current status; only
// then do we update Stripe and the local record.
func (s *SubscriptionCommandService) Cancel(cmd commands.CancelSubscriptionCommand) (*entities.Subscription, error) {
	subscription, err := s.subscriptions.FindByID(cmd.SubscriptionID)
	if err != nil {
		return nil, err
	}
	// The repository returns (nil, nil) when nothing is found, so we translate a
	// missing subscription into the standard "not found" error here.
	if subscription == nil {
		return nil, gorm.ErrRecordNotFound
	}
	if err = s.manager.CanCancel(subscription.Status); err != nil {
		return nil, err
	}
	// Only call Stripe if this subscription is actually linked to it. The result
	// is ignored on purpose: a local cancellation should still proceed even if
	// the remote call fails.
	if subscription.StripeSubscriptionID != nil {
		_ = s.stripe.CancelSubscription(*subscription.StripeSubscriptionID)
	}
	now := time.Now().UTC()
	subscription.Status = valueobjects.StatusCancelled
	subscription.EndDate = &now
	if err = s.subscriptions.Update(subscription); err != nil {
		return nil, err
	}
	s.publish(integrationevents.EventTypeSubscriptionCancelled, subscription, nil, nil)
	return subscription, nil
}

// ChangePlan switches an existing subscription to a different plan. It validates
// the transition, checks the new plan, updates Stripe, then persists. The new
// status becomes PENDING_RENEWAL because the billing change takes effect on the
// next cycle.
func (s *SubscriptionCommandService) ChangePlan(cmd commands.ChangePlanCommand) (*entities.Subscription, error) {
	subscription, err := s.subscriptions.FindByID(cmd.SubscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, gorm.ErrRecordNotFound
	}
	if err = s.manager.CanChangePlan(subscription.Status); err != nil {
		return nil, err
	}
	plan, err := s.plans.FindByID(cmd.NewPlanID)
	if err != nil {
		return nil, err
	}
	if plan == nil || !plan.Active {
		return nil, errors.New("new plan is not available")
	}
	targetPriceID, err := extractStripePriceID(*plan)
	if err != nil {
		return nil, err
	}
	if subscription.StripeSubscriptionID != nil {
		_ = s.stripe.ChangePlan(*subscription.StripeSubscriptionID, targetPriceID)
	}
	previousPlanID := subscription.PlanID
	subscription.PlanID = cmd.NewPlanID
	subscription.Status = valueobjects.StatusPendingRenewal
	if err = s.subscriptions.Update(subscription); err != nil {
		return nil, err
	}
	s.publish(integrationevents.EventTypeSubscriptionPlanChanged, subscription, plan, map[string]any{
		"previousPlanId": previousPlanID,
		"newPlanId":      cmd.NewPlanID,
	})
	s.publish(integrationevents.EventTypeSubscriptionRenewalRequested, subscription, plan, map[string]any{
		"previousPlanId": previousPlanID,
		"newPlanId":      cmd.NewPlanID,
		"reason":         "plan_change",
	})
	return subscription, nil
}

// extractStripePriceID searches a plan's feature list for the STRIPE_PRICE_ID
// entry and returns its value. Storing the price id as a "feature" keeps the
// plan model flexible (any key/value can be attached) at the cost of this small
// lookup. strings.EqualFold compares the code case-insensitively.
func extractStripePriceID(plan entities.SubscriptionPlan) (string, error) {
	for _, f := range plan.PlanFeatures {
		if strings.EqualFold(strings.TrimSpace(f.FeatureCode), "STRIPE_PRICE_ID") {
			value := strings.TrimSpace(f.FeatureValue)
			if value != "" {
				return value, nil
			}
		}
	}
	return "", errors.New("plan is missing STRIPE_PRICE_ID in plan features")
}

// publish sends a normalized event envelope to the grouped subscriptions topic.
func (s *SubscriptionCommandService) publish(eventType string, subscription *entities.Subscription, plan *entities.SubscriptionPlan, extraData map[string]any) {
	if s.events == nil {
		return
	}
	enrichedData := make(map[string]any, len(extraData)+5)
	for key, value := range extraData {
		enrichedData[key] = value
	}
	if plan != nil {
		enrichedData["planName"] = plan.Name
		enrichedData["billingPeriod"] = plan.BillingPeriod
		enrichedData["currency"] = plan.Currency
		enrichedData["amount"] = plan.Price
		enrichedData["requiresPayment"] = plan.Price > 0
	}
	payload, err := integrationevents.MarshalSubscriptionEvent(eventType, time.Now().UTC(), subscription, enrichedData)
	if err != nil {
		return
	}
	_ = s.events.Publish(s.topics.Events, payload)
}
