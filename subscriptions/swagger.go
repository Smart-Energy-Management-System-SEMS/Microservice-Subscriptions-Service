package subscriptions

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Subscriptions Service API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    body { margin: 0; background: #f4f7fb; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/swagger/doc.json',
      dom_id: '#swagger-ui',
      deepLinking: true,
      displayRequestDuration: true,
      tryItOutEnabled: true,
      persistAuthorization: true
    });
  </script>
</body>
</html>
`

func RegisterDocsRoutes(r *gin.Engine) {
	r.GET("/swagger", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	r.GET("/swagger/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	r.GET("/swagger/index.html", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIHTML))
	})
	r.GET("/swagger/doc.json", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, buildOpenAPISpec())
	})
}

func buildOpenAPISpec() gin.H {
	return gin.H{
		"openapi": "3.0.3",
		"info": gin.H{
			"title":       "Subscriptions Service API",
			"description": "Interactive documentation for SEMS subscriptions endpoints, including health checks, plan management, subscriptions, and the Stripe webhook.",
			"version":     "1.0.0",
		},
		"servers": []gin.H{
			{"url": "/"},
		},
		"paths": gin.H{
			"/health": gin.H{
				"get": gin.H{
					"tags":        []string{"Health"},
					"summary":     "Basic health check",
					"description": "Quick service probe.",
					"responses": gin.H{
						"200": jsonResponse("Service is healthy", "HealthResponse"),
					},
				},
			},
			"/api/v1/health": gin.H{
				"get": gin.H{
					"tags":        []string{"Health"},
					"summary":     "Versioned health check",
					"description": "Versioned probe used by clients and gateways.",
					"responses": gin.H{
						"200": jsonResponse("Service is healthy", "HealthResponse"),
					},
				},
			},
			"/api/v1/subscription-plans": gin.H{
				"get": gin.H{
					"tags":        []string{"Plans"},
					"summary":     "List subscription plans",
					"description": "Returns the fixed catalog of available subscription plans.",
					"responses": gin.H{
						"200": gin.H{
							"description": "Plans retrieved successfully",
							"content": gin.H{
								"application/json": gin.H{
									"schema": gin.H{
										"type":  "array",
										"items": refSchema("SubscriptionPlanResponse"),
									},
								},
							},
						},
						"500": jsonResponse("Unexpected server error", "ErrorResponse"),
					},
				},
			},
			"/api/v1/subscription-plans/{planId}": gin.H{
				"parameters": []gin.H{pathParam("planId", "Plan identifier")},
				"get": gin.H{
					"tags":        []string{"Plans"},
					"summary":     "Get plan by id",
					"description": "Returns a single plan.",
					"responses": gin.H{
						"200": jsonResponse("Plan retrieved", "SubscriptionPlanResponse"),
						"404": jsonResponse("Plan not found", "ErrorResponse"),
						"500": jsonResponse("Unexpected server error", "ErrorResponse"),
					},
				},
			},
			"/api/v1/subscriptions": gin.H{
				"post": gin.H{
					"tags":        []string{"Subscriptions"},
					"summary":     "Create subscription",
					"description": "Creates a subscription for a user and publishes the corresponding event.",
					"requestBody": gin.H{
						"required": true,
						"content": gin.H{
							"application/json": gin.H{
								"schema": refSchema("CreateSubscriptionRequest"),
								"example": gin.H{
									"user_id":            "user-001",
									"plan_id":            "plan-pro",
									"stripe_customer_id": "cus_test_123",
								},
							},
						},
					},
					"responses": gin.H{
						"201": jsonResponse("Subscription created", "SubscriptionResponse"),
						"400": jsonResponse("Invalid request", "ErrorResponse"),
					},
				},
			},
			"/api/v1/subscriptions/{subscriptionId}": gin.H{
				"parameters": []gin.H{pathParam("subscriptionId", "Subscription identifier")},
				"get": gin.H{
					"tags":        []string{"Subscriptions"},
					"summary":     "Get subscription by id",
					"description": "Returns a single subscription.",
					"responses": gin.H{
						"200": jsonResponse("Subscription retrieved", "SubscriptionResponse"),
						"404": jsonResponse("Subscription not found", "ErrorResponse"),
						"500": jsonResponse("Unexpected server error", "ErrorResponse"),
					},
				},
			},
			"/api/v1/subscriptions/users/{userId}": gin.H{
				"parameters": []gin.H{pathParam("userId", "User identifier")},
				"get": gin.H{
					"tags":        []string{"Subscriptions"},
					"summary":     "List subscriptions by user",
					"description": "Returns all subscriptions for a given user.",
					"responses": gin.H{
						"200": gin.H{
							"description": "Subscriptions retrieved successfully",
							"content": gin.H{
								"application/json": gin.H{
									"schema": gin.H{
										"type":  "array",
										"items": refSchema("SubscriptionResponse"),
									},
								},
							},
						},
						"500": jsonResponse("Unexpected server error", "ErrorResponse"),
					},
				},
			},
			"/api/v1/subscriptions/{subscriptionId}/cancel": gin.H{
				"parameters": []gin.H{pathParam("subscriptionId", "Subscription identifier")},
				"patch": gin.H{
					"tags":        []string{"Subscriptions"},
					"summary":     "Cancel subscription",
					"description": "Cancels an existing subscription.",
					"responses": gin.H{
						"200": jsonResponse("Subscription cancelled", "SubscriptionResponse"),
						"400": jsonResponse("Invalid request", "ErrorResponse"),
						"404": jsonResponse("Subscription not found", "ErrorResponse"),
					},
				},
			},
			"/api/v1/subscriptions/{subscriptionId}/change-plan": gin.H{
				"parameters": []gin.H{pathParam("subscriptionId", "Subscription identifier")},
				"patch": gin.H{
					"tags":        []string{"Subscriptions"},
					"summary":     "Change subscription plan",
					"description": "Requests a plan change and emits the plan-change and renewal-requested events.",
					"requestBody": gin.H{
						"required": true,
						"content": gin.H{
							"application/json": gin.H{
								"schema": refSchema("ChangePlanRequest"),
								"example": gin.H{
									"new_plan_id": "plan-plus",
								},
							},
						},
					},
					"responses": gin.H{
						"200": jsonResponse("Subscription updated", "SubscriptionResponse"),
						"400": jsonResponse("Invalid request", "ErrorResponse"),
						"404": jsonResponse("Subscription not found", "ErrorResponse"),
					},
				},
			},
			"/api/v1/webhooks/stripe": gin.H{
				"post": gin.H{
					"tags":        []string{"Webhooks"},
					"summary":     "Receive Stripe webhook",
					"description": "Receives raw Stripe webhook events. For real tests, send the exact payload Stripe signs and include the Stripe-Signature header.",
					"parameters": []gin.H{
						{
							"name":        "Stripe-Signature",
							"in":          "header",
							"required":    true,
							"description": "Stripe signature header generated for the raw payload.",
							"schema": gin.H{
								"type": "string",
							},
						},
					},
					"requestBody": gin.H{
						"required": true,
						"content": gin.H{
							"application/json": gin.H{
								"schema": gin.H{
									"type": "object",
								},
								"example": gin.H{
									"id":   "evt_test_webhook",
									"type": "customer.subscription.created",
								},
							},
						},
					},
					"responses": gin.H{
						"200": jsonResponse("Webhook accepted", "WebhookAckResponse"),
						"400": jsonResponse("Invalid webhook payload or signature", "ErrorResponse"),
						"503": jsonResponse("Webhook is not configured", "ErrorResponse"),
					},
				},
			},
		},
		"components": gin.H{
			"schemas": gin.H{
				"HealthResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"status": gin.H{"type": "string", "example": "ok"},
					},
				},
				"ErrorResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"error": gin.H{"type": "string", "example": "plan not found"},
					},
				},
				"PlanFeatureRequest": gin.H{
					"type": "object",
					"properties": gin.H{
						"feature_code":  gin.H{"type": "string", "example": "STRIPE_PRICE_ID"},
						"feature_name":  gin.H{"type": "string", "example": "Stripe price id"},
						"feature_value": gin.H{"type": "string", "example": "price_1Tbx5vJTjqngaxe90d0NecBm"},
					},
				},
				"CreatePlanRequest": gin.H{
					"type":     "object",
					"required": []string{"name", "price", "currency", "billing_period"},
					"properties": gin.H{
						"name":           gin.H{"type": "string", "example": "Pro"},
						"description":    gin.H{"type": "string", "example": "Plan con analytics y automatizaciones"},
						"price":          gin.H{"type": "number", "format": "double", "example": 59.9},
						"currency":       gin.H{"type": "string", "example": "pen"},
						"billing_period": gin.H{"type": "string", "example": "monthly"},
						"features": gin.H{
							"type":  "array",
							"items": refSchema("PlanFeatureRequest"),
						},
					},
				},
				"UpdatePlanRequest": gin.H{
					"type":     "object",
					"required": []string{"name", "price", "currency", "billing_period"},
					"properties": gin.H{
						"name":           gin.H{"type": "string", "example": "Pro Annual"},
						"description":    gin.H{"type": "string", "example": "Plan pro con cobro anual"},
						"price":          gin.H{"type": "number", "format": "double", "example": 599.0},
						"currency":       gin.H{"type": "string", "example": "pen"},
						"billing_period": gin.H{"type": "string", "example": "annual"},
					},
				},
				"PlanFeatureResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"FeatureID":    gin.H{"type": "string", "example": "feat-001"},
						"PlanID":       gin.H{"type": "string", "example": "plan-pro"},
						"FeatureCode":  gin.H{"type": "string", "example": "STRIPE_PRICE_ID"},
						"FeatureName":  gin.H{"type": "string", "example": "Stripe price id"},
						"FeatureValue": gin.H{"type": "string", "example": "price_1Tbx5vJTjqngaxe90d0NecBm"},
						"CreatedAt":    gin.H{"type": "string", "format": "date-time"},
					},
				},
				"SubscriptionPlanResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"PlanID":        gin.H{"type": "string", "example": "plan-pro"},
						"Name":          gin.H{"type": "string", "example": "Pro"},
						"Description":   gin.H{"type": "string", "example": "Plan con analytics y automatizaciones"},
						"Price":         gin.H{"type": "number", "format": "double", "example": 59.9},
						"Currency":      gin.H{"type": "string", "example": "pen"},
						"BillingPeriod": gin.H{"type": "string", "example": "monthly"},
						"Active":        gin.H{"type": "boolean", "example": true},
						"CreatedAt":     gin.H{"type": "string", "format": "date-time"},
						"PlanFeatures": gin.H{
							"type":  "array",
							"items": refSchema("PlanFeatureResponse"),
						},
					},
				},
				"CreateSubscriptionRequest": gin.H{
					"type":     "object",
					"required": []string{"user_id", "plan_id"},
					"properties": gin.H{
						"user_id":            gin.H{"type": "string", "example": "user-001"},
						"plan_id":            gin.H{"type": "string", "example": "plan-pro"},
						"stripe_customer_id": gin.H{"type": "string", "example": "cus_test_123"},
					},
				},
				"ChangePlanRequest": gin.H{
					"type":     "object",
					"required": []string{"new_plan_id"},
					"properties": gin.H{
						"new_plan_id": gin.H{"type": "string", "example": "plan-plus"},
					},
				},
				"SubscriptionResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"SubscriptionID":       gin.H{"type": "string", "example": "sub-001"},
						"UserID":               gin.H{"type": "string", "example": "user-001"},
						"PlanID":               gin.H{"type": "string", "example": "plan-pro"},
						"Status":               gin.H{"type": "string", "example": "ACTIVE"},
						"StartDate":            gin.H{"type": "string", "format": "date-time"},
						"EndDate":              gin.H{"type": "string", "format": "date-time", "nullable": true},
						"StripeSubscriptionID": gin.H{"type": "string", "nullable": true, "example": "sub_stripe_123"},
						"CreatedAt":            gin.H{"type": "string", "format": "date-time"},
					},
				},
				"WebhookAckResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"received": gin.H{"type": "boolean", "example": true},
					},
				},
			},
		},
	}
}

func jsonResponse(description, schema string) gin.H {
	return gin.H{
		"description": description,
		"content": gin.H{
			"application/json": gin.H{
				"schema": refSchema(schema),
			},
		},
	}
}

func refSchema(name string) gin.H {
	return gin.H{"$ref": "#/components/schemas/" + name}
}

func pathParam(name, description string) gin.H {
	return gin.H{
		"name":        name,
		"in":          "path",
		"required":    true,
		"description": description,
		"schema": gin.H{
			"type": "string",
		},
	}
}
