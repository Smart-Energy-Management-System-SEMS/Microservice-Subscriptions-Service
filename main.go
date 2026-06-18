package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"microservice-subscriptions-service/subscriptions"
	"microservice-subscriptions-service/subscriptions/application/commandservices"
	"microservice-subscriptions-service/subscriptions/application/eventhandlers"
	"microservice-subscriptions-service/subscriptions/application/queryservices"
	appconfig "microservice-subscriptions-service/subscriptions/infrastructure/configuration"
	kafkainfra "microservice-subscriptions-service/subscriptions/infrastructure/messaging/kafka"
	stripeinfra "microservice-subscriptions-service/subscriptions/infrastructure/payments/stripe"
	dbconfig "microservice-subscriptions-service/subscriptions/infrastructure/persistence/gorm/configuration"
	gormrepo "microservice-subscriptions-service/subscriptions/infrastructure/persistence/gorm/repositories"
	"microservice-subscriptions-service/subscriptions/interfaces/acl"
	"microservice-subscriptions-service/subscriptions/interfaces/rest/controllers"
)

func main() {
	cfg := appconfig.Load()
	if cfg.KafkaEnabled {
		log.Printf(
			"Kafka publisher config: brokers=%s security_protocol=%s sasl_mechanism=%s username=%q",
			strings.Join(cfg.KafkaBrokers, ","),
			cfg.KafkaSecurityProto,
			cfg.KafkaSASLMechanism,
			cfg.KafkaUsername,
		)
	}
	db, err := dbconfig.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(fmt.Errorf("database connection failed: %w", err))
	}

	if err = dbconfig.AutoMigrate(db); err != nil {
		log.Fatal(fmt.Errorf("database migration failed: %w", err))
	}
	if err = dbconfig.SeedDefaultPlans(db, cfg); err != nil {
		log.Fatal(fmt.Errorf("default plans seed failed: %w", err))
	}

	planRepo := gormrepo.NewSubscriptionPlanRepository(db)
	subRepo := gormrepo.NewSubscriptionRepository(db)
	publisher := kafkainfra.NewPublisher(kafkainfra.Config{
		Enabled:          cfg.KafkaEnabled,
		Brokers:          cfg.KafkaBrokers,
		ClientID:         cfg.KafkaClientID,
		Username:         cfg.KafkaUsername,
		Password:         cfg.KafkaPassword,
		SecurityProtocol: cfg.KafkaSecurityProto,
		SASLMechanism:    cfg.KafkaSASLMechanism,
		CACert:           cfg.KafkaCACert,
		CACertPath:       cfg.KafkaCACertPath,
	})
	stripeAdapter := stripeinfra.NewAdapter(cfg.StripeSecretKey)
	stripeServiceACL := acl.NewStripeServiceACL(stripeAdapter)

	planCommand := commandservices.NewPlanCommandService(planRepo)
	planQuery := queryservices.NewPlanQueryService(planRepo)
	subscriptionCommand := commandservices.NewSubscriptionCommandService(
		subRepo,
		planRepo,
		stripeServiceACL,
		publisher,
		commandservices.SubscriptionTopics{
			Events: cfg.KafkaTopicSubscriptionsEvents,
		},
	)
	subscriptionQuery := queryservices.NewSubscriptionQueryService(subRepo)
	stripeWebhookHandler := eventhandlers.NewStripeWebhookHandler(
		subRepo,
		publisher,
		eventhandlers.StripeWebhookTopics{
			Events: cfg.KafkaTopicSubscriptionsEvents,
		},
	)

	controller := controllers.NewSubscriptionController(planCommand, planQuery, subscriptionCommand, subscriptionQuery, stripeWebhookHandler, cfg.StripeWebhookSecret)
	r := gin.Default()
	r.SetTrustedProxies(nil)
	allowedOrigins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		trimmedOrigin := strings.TrimSpace(origin)
		if trimmedOrigin != "" {
			allowedOrigins[trimmedOrigin] = struct{}{}
		}
	}
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowedOrigins[origin]; ok {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Origin, Accept, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	subscriptions.RegisterRoutes(r, controller)

	if err = r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
