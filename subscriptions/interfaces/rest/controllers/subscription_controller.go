package controllers

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"microservice-subscriptions-service/subscriptions/application/commandservices"
	"microservice-subscriptions-service/subscriptions/application/eventhandlers"
	"microservice-subscriptions-service/subscriptions/application/queryservices"
	"microservice-subscriptions-service/subscriptions/domain/model/commands"
	"microservice-subscriptions-service/subscriptions/interfaces/rest/resources"
	"microservice-subscriptions-service/subscriptions/interfaces/rest/transform"
)

// Package controllers is the REST entry point (interfaces layer) built on the
// Gin web framework. A controller's only job is to handle HTTP: read the
// request, call the right application service, and turn the result or error into
// an HTTP response. It contains no business logic.

// SubscriptionController handles both plan and subscription endpoints. It groups
// command services (writes) and query services (reads) — the CQRS split — plus
// the Stripe webhook handler and its secret.
type SubscriptionController struct {
	planCommand         *commandservices.PlanCommandService
	planQuery           *queryservices.PlanQueryService
	subscriptionCommand *commandservices.SubscriptionCommandService
	subscriptionQuery   *queryservices.SubscriptionQueryService
	stripeWebhook       *eventhandlers.StripeWebhookHandler
	stripeWebhookSecret string
}

// NewSubscriptionController injects all the services the handlers need.
func NewSubscriptionController(planCommand *commandservices.PlanCommandService, planQuery *queryservices.PlanQueryService, subscriptionCommand *commandservices.SubscriptionCommandService, subscriptionQuery *queryservices.SubscriptionQueryService, stripeWebhook *eventhandlers.StripeWebhookHandler, stripeWebhookSecret string) *SubscriptionController {
	return &SubscriptionController{planCommand: planCommand, planQuery: planQuery, subscriptionCommand: subscriptionCommand, subscriptionQuery: subscriptionQuery, stripeWebhook: stripeWebhook, stripeWebhookSecret: stripeWebhookSecret}
}

// GetPlans returns every plan. Read handlers like this are short: call the query
// service and respond. internalError maps an unexpected failure to HTTP 500.
func (c *SubscriptionController) GetPlans(ctx *gin.Context) {
	plans, err := c.planQuery.FindAll()
	if err != nil {
		internalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, plans)
}

// GetPlanByID reads the "planId" path parameter and returns that plan, or 404
// when the query service reports it as missing (a nil result).
func (c *SubscriptionController) GetPlanByID(ctx *gin.Context) {
	plan, err := c.planQuery.FindByID(ctx.Param("planId"))
	if err != nil {
		internalError(ctx, err)
		return
	}
	if plan == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}
	ctx.JSON(http.StatusOK, plan)
}

// CreatePlan shows the typical write-handler flow: bind+validate the JSON body,
// map the request to a domain command, call the command service, and respond
// with 201 Created. ShouldBindJSON returns an error if the body is malformed.
func (c *SubscriptionController) CreatePlan(ctx *gin.Context) {
	var req resources.CreatePlanRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		badRequest(ctx, err)
		return
	}
	plan, err := c.planCommand.Create(transform.ToCreatePlanCommand(req))
	if err != nil {
		badRequest(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, plan)
}

// UpdatePlan adds one more idea on top of CreatePlan: it distinguishes error
// kinds. errors.Is checks whether the failure is specifically "record not found"
// (-> 404); anything else is treated as a bad request (-> 400). Mapping domain
// errors to the right HTTP status is a core responsibility of the controller.
func (c *SubscriptionController) UpdatePlan(ctx *gin.Context) {
	var req resources.UpdatePlanRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		badRequest(ctx, err)
		return
	}
	plan, err := c.planCommand.Update(transform.ToUpdatePlanCommand(ctx.Param("planId"), req))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		badRequest(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, plan)
}

// DeactivatePlan is a soft delete: it does not remove the plan, it marks it
// inactive. On success it returns 204 No Content (a successful response with an
// empty body), which is the conventional answer for this kind of action.
func (c *SubscriptionController) DeactivatePlan(ctx *gin.Context) {
	if err := c.planCommand.Deactivate(ctx.Param("planId")); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		internalError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *SubscriptionController) GetSubscriptionByID(ctx *gin.Context) {
	sub, err := c.subscriptionQuery.FindByID(ctx.Param("subscriptionId"))
	if err != nil {
		internalError(ctx, err)
		return
	}
	if sub == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}
	ctx.JSON(http.StatusOK, sub)
}

func (c *SubscriptionController) GetSubscriptionsByUserID(ctx *gin.Context) {
	subs, err := c.subscriptionQuery.FindByUserID(ctx.Param("userId"))
	if err != nil {
		internalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, subs)
}

// CreateSubscription is the main subscription write endpoint. Same pattern as
// CreatePlan: validate the body, transform it into a command, delegate to the
// command service, and respond 201 on success.
func (c *SubscriptionController) CreateSubscription(ctx *gin.Context) {
	var req resources.CreateSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		badRequest(ctx, err)
		return
	}
	sub, err := c.subscriptionCommand.Create(transform.ToCreateSubscriptionCommand(req))
	if err != nil {
		badRequest(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, sub)
}

func (c *SubscriptionController) CancelSubscription(ctx *gin.Context) {
	sub, err := c.subscriptionCommand.Cancel(commands.CancelSubscriptionCommand{SubscriptionID: ctx.Param("subscriptionId")})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		badRequest(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, sub)
}

func (c *SubscriptionController) ChangePlan(ctx *gin.Context) {
	var req resources.ChangePlanRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		badRequest(ctx, err)
		return
	}
	sub, err := c.subscriptionCommand.ChangePlan(commands.ChangePlanCommand{SubscriptionID: ctx.Param("subscriptionId"), NewPlanID: req.NewPlanID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		badRequest(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, sub)
}

// StripeWebhook receives the callbacks Stripe sends when something happens on
// its side. A few details are specific to webhooks:
//   - If the webhook is not configured, we answer 503 so it is clear the feature
//     is unavailable rather than silently failing.
//   - The "Stripe-Signature" header is required; it is used to prove the request
//     genuinely came from Stripe and was not forged.
//   - We read the RAW body with io.ReadAll. The exact bytes matter because the
//     signature is computed over them; parsing into a struct first would break
//     verification.
// The actual signature check and event processing happen in the webhook handler.
func (c *SubscriptionController) StripeWebhook(ctx *gin.Context) {
	if c.stripeWebhook == nil || c.stripeWebhookSecret == "" {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "stripe webhook is not configured"})
		return
	}

	signature := ctx.GetHeader("Stripe-Signature")
	if signature == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing Stripe-Signature header"})
		return
	}

	payload, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err = c.stripeWebhook.Handle(payload, signature, c.stripeWebhookSecret); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"received": true})
}

// badRequest and internalError are tiny helpers that keep the handlers above
// short and make every error response consistent. badRequest -> HTTP 400 (the
// client sent something invalid); internalError -> HTTP 500 (something failed on
// our side).
func badRequest(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
func internalError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
