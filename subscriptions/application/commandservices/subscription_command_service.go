package commandservices

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"microservice-subscriptions-service/subscriptions/application/outboundservices"
	"microservice-subscriptions-service/subscriptions/domain/model/commands"
	"microservice-subscriptions-service/subscriptions/domain/model/entities"
	"microservice-subscriptions-service/subscriptions/domain/model/valueobjects"
	domainrepo "microservice-subscriptions-service/subscriptions/domain/repositories"
	"microservice-subscriptions-service/subscriptions/domain/services"
)

type SubscriptionCommandService struct {
	subscriptions domainrepo.SubscriptionRepository
	plans         domainrepo.SubscriptionPlanRepository
	stripe        outboundservices.StripeService
	events        outboundservices.EventPublisher
	manager       *services.SubscriptionManager
	topics        SubscriptionTopics
}

type SubscriptionTopics struct {
	Created     string
	Cancelled   string
	PlanChanged string
}

func NewSubscriptionCommandService(subscriptions domainrepo.SubscriptionRepository, plans domainrepo.SubscriptionPlanRepository, stripe outboundservices.StripeService, events outboundservices.EventPublisher, topics SubscriptionTopics) *SubscriptionCommandService {
	if strings.TrimSpace(topics.Created) == "" {
		topics.Created = "SubscriptionCreated"
	}
	if strings.TrimSpace(topics.Cancelled) == "" {
		topics.Cancelled = "SubscriptionCancelled"
	}
	if strings.TrimSpace(topics.PlanChanged) == "" {
		topics.PlanChanged = "SubscriptionPlanChanged"
	}
	return &SubscriptionCommandService{subscriptions: subscriptions, plans: plans, stripe: stripe, events: events, manager: services.NewSubscriptionManager(), topics: topics}
}

func (s *SubscriptionCommandService) Create(cmd commands.CreateSubscriptionCommand) (*entities.Subscription, error) {
	if cmd.UserID == "" || cmd.PlanID == "" {
		return nil, errors.New("user_id and plan_id are required")
	}
	plan, err := s.plans.FindByID(cmd.PlanID)
	if err != nil {
		return nil, err
	}
	if plan == nil || !plan.Active {
		return nil, errors.New("plan is not available")
	}
	stripePriceID, err := extractStripePriceID(*plan)
	if err != nil {
		return nil, err
	}
	subscription := &entities.Subscription{SubscriptionID: uuid.NewString(), UserID: cmd.UserID, PlanID: cmd.PlanID, Status: valueobjects.StatusActive, StartDate: time.Now().UTC(), CreatedAt: time.Now().UTC()}
	stripeID, err := s.stripe.CreateSubscription(*subscription, cmd.StripeCustomerID, stripePriceID)
	if err == nil && stripeID != "" {
		subscription.StripeSubscriptionID = &stripeID
	}
	if err = s.subscriptions.Create(subscription); err != nil {
		return nil, err
	}
	s.publish(s.topics.Created, subscription)
	return subscription, nil
}

func (s *SubscriptionCommandService) Cancel(cmd commands.CancelSubscriptionCommand) (*entities.Subscription, error) {
	subscription, err := s.subscriptions.FindByID(cmd.SubscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, gorm.ErrRecordNotFound
	}
	if err = s.manager.CanCancel(subscription.Status); err != nil {
		return nil, err
	}
	if subscription.StripeSubscriptionID != nil {
		_ = s.stripe.CancelSubscription(*subscription.StripeSubscriptionID)
	}
	now := time.Now().UTC()
	subscription.Status = valueobjects.StatusCancelled
	subscription.EndDate = &now
	if err = s.subscriptions.Update(subscription); err != nil {
		return nil, err
	}
	s.publish(s.topics.Cancelled, subscription)
	return subscription, nil
}

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
	subscription.PlanID = cmd.NewPlanID
	subscription.Status = valueobjects.StatusPendingRenewal
	if err = s.subscriptions.Update(subscription); err != nil {
		return nil, err
	}
	s.publish(s.topics.PlanChanged, subscription)
	return subscription, nil
}

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

func (s *SubscriptionCommandService) publish(topic string, subscription *entities.Subscription) {
	if s.events == nil {
		return
	}
	payload, err := json.Marshal(subscription)
	if err != nil {
		return
	}
	_ = s.events.Publish(topic, payload)
}
