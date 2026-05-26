package subscriptions

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"microservice-subscriptions-service/subscriptions/application/commandservices"
	"microservice-subscriptions-service/subscriptions/application/queryservices"
	appconfig "microservice-subscriptions-service/subscriptions/infrastructure/configuration"
	kafkainfra "microservice-subscriptions-service/subscriptions/infrastructure/messaging/kafka"
	stripeinfra "microservice-subscriptions-service/subscriptions/infrastructure/payments/stripe"
	dbconfig "microservice-subscriptions-service/subscriptions/infrastructure/persistence/gorm/configuration"
	gormrepo "microservice-subscriptions-service/subscriptions/infrastructure/persistence/gorm/repositories"
	"microservice-subscriptions-service/subscriptions/interfaces/rest/controllers"
)

func Start() error {
	cfg := appconfig.Load()
	db, err := dbconfig.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	if err = dbconfig.AutoMigrate(db); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	planRepo := gormrepo.NewSubscriptionPlanRepository(db)
	subRepo := gormrepo.NewSubscriptionRepository(db)
	publisher := kafkainfra.NewPublisher(cfg.KafkaBrokers, cfg.KafkaClientID)
	stripeAdapter := stripeinfra.NewAdapter(cfg.StripeSecretKey)

	planCommand := commandservices.NewPlanCommandService(planRepo)
	planQuery := queryservices.NewPlanQueryService(planRepo)
	subscriptionCommand := commandservices.NewSubscriptionCommandService(subRepo, planRepo, stripeAdapter, publisher)
	subscriptionQuery := queryservices.NewSubscriptionQueryService(subRepo)

	controller := controllers.NewSubscriptionController(planCommand, planQuery, subscriptionCommand, subscriptionQuery)
	r := gin.Default()
	RegisterRoutes(r, controller)

	return r.Run(":" + cfg.ServerPort)
}
