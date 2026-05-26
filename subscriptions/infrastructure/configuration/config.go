package configuration

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	ServerPort           string
	DatabaseURL          string
	StripePublishableKey string
	StripeSecretKey      string
	StripeWebhookSecret  string
	StripePriceFree      string
	StripePricePlus      string
	StripePricePro       string
	KafkaBrokers         []string
	KafkaClientID        string
}

func Load() AppConfig {
	_ = godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = buildPostgresURLFromEnv()
	}

	return AppConfig{
		ServerPort:           getEnv("SERVER_PORT", "8081"),
		DatabaseURL:          databaseURL,
		StripePublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		StripeSecretKey:      os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:  os.Getenv("STRIPE_WEBHOOK_SECRET"),
		StripePriceFree:      os.Getenv("STRIPE_PRICE_FREE"),
		StripePricePlus:      os.Getenv("STRIPE_PRICE_PLUS"),
		StripePricePro:       os.Getenv("STRIPE_PRICE_PRO"),
		KafkaBrokers:         splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaClientID:        getEnv("KAFKA_CLIENT_ID", "subscriptions-service"),
	}
}

func buildPostgresURLFromEnv() string {
	host := os.Getenv("POSTGRES_HOST")
	port := getEnv("POSTGRES_PORT", "5432")
	dbName := os.Getenv("POSTGRES_DB")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	sslMode := getEnv("POSTGRES_SSLMODE", "require")
	channelBinding := getEnv("POSTGRES_CHANNEL_BINDING", "require")

	if host == "" || dbName == "" || user == "" || password == "" {
		return ""
	}

	query := url.Values{}
	query.Set("sslmode", sslMode)
	query.Set("channelBinding", channelBinding)

	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?%s",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		dbName,
		query.Encode(),
	)
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
