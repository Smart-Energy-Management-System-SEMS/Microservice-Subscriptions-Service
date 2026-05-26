package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"microservice-subscriptions-service/subscriptions/application/commandservices"
	"microservice-subscriptions-service/subscriptions/application/queryservices"
	"microservice-subscriptions-service/subscriptions/domain/model/commands"
	"microservice-subscriptions-service/subscriptions/interfaces/rest/resources"
	"microservice-subscriptions-service/subscriptions/interfaces/rest/transform"
)

type SubscriptionController struct {
	planCommand         *commandservices.PlanCommandService
	planQuery           *queryservices.PlanQueryService
	subscriptionCommand *commandservices.SubscriptionCommandService
	subscriptionQuery   *queryservices.SubscriptionQueryService
}

func NewSubscriptionController(planCommand *commandservices.PlanCommandService, planQuery *queryservices.PlanQueryService, subscriptionCommand *commandservices.SubscriptionCommandService, subscriptionQuery *queryservices.SubscriptionQueryService) *SubscriptionController {
	return &SubscriptionController{planCommand: planCommand, planQuery: planQuery, subscriptionCommand: subscriptionCommand, subscriptionQuery: subscriptionQuery}
}

func (c *SubscriptionController) GetPlans(ctx *gin.Context) {
	plans, err := c.planQuery.FindAll()
	if err != nil {
		internalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, plans)
}

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

func badRequest(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
func internalError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
