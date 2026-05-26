package configuration

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	ServerPort          string
	DatabaseURL         string
	StripeSecretKey     string
	StripeWebhookSecret string
	KafkaBrokers        []string
	KafkaClientID       string
}

func Load() AppConfig {
	_ = godotenv.Load()
	return AppConfig{
		ServerPort:          getEnv("SERVER_PORT", "8081"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		KafkaBrokers:        splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaClientID:       getEnv("KAFKA_CLIENT_ID", "subscriptions-service"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			clean = append(clean, p)
		}
	}
	return clean
}
