package subscriptions

import (
	"github.com/gin-gonic/gin"
	"microservice-subscriptions-service/subscriptions/interfaces/rest/controllers"
)

func RegisterRoutes(r *gin.Engine, c *controllers.SubscriptionController) {
	RegisterDocsRoutes(r)

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/api/v1/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		plans := v1.Group("/subscription-plans")
		{
			plans.GET("", c.GetPlans)
			plans.GET("/:planId", c.GetPlanByID)
		}

		subs := v1.Group("/subscriptions")
		{
			subs.GET("/:subscriptionId", c.GetSubscriptionByID)
			subs.GET("/users/:userId", c.GetSubscriptionsByUserID)
			subs.POST("", c.CreateSubscription)
			subs.PATCH("/:subscriptionId/cancel", c.CancelSubscription)
			subs.PATCH("/:subscriptionId/change-plan", c.ChangePlan)
		}

		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("/stripe", c.StripeWebhook)
		}
	}
}
