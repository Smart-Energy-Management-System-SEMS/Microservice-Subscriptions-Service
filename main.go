package main

import (
	"fmt"
	"log"

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
	publisher := kafkainfra.NewPublisher(cfg.KafkaBrokers, cfg.KafkaClientID)
	stripeAdapter := stripeinfra.NewAdapter(cfg.StripeSecretKey)
	stripeServiceACL := acl.NewStripeServiceACL(stripeAdapter)

	planCommand := commandservices.NewPlanCommandService(planRepo)
	planQuery := queryservices.NewPlanQueryService(planRepo)
	subscriptionCommand := commandservices.NewSubscriptionCommandService(subRepo, planRepo, stripeServiceACL, publisher)
	subscriptionQuery := queryservices.NewSubscriptionQueryService(subRepo)
	stripeWebhookHandler := eventhandlers.NewStripeWebhookHandler(subRepo, publisher)

	controller := controllers.NewSubscriptionController(planCommand, planQuery, subscriptionCommand, subscriptionQuery, stripeWebhookHandler, cfg.StripeWebhookSecret)
	r := gin.Default()
	subscriptions.RegisterRoutes(r, controller)

	if err = r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
