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
	KafkaEnabled         bool
	KafkaBootstrapServer string
	KafkaBrokers         []string
	KafkaClientID        string
	KafkaUsername        string
	KafkaPassword        string
	KafkaSecurityProto   string
	KafkaSASLMechanism   string
	KafkaCACert          string
	KafkaCACertPath      string
}

func Load() AppConfig {
	_ = godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = buildPostgresURLFromEnv()
	}

	return AppConfig{
		ServerPort:           getServerPort(),
		DatabaseURL:          databaseURL,
		StripePublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		StripeSecretKey:      os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:  os.Getenv("STRIPE_WEBHOOK_SECRET"),
		StripePriceFree:      os.Getenv("STRIPE_PRICE_FREE"),
		StripePricePlus:      os.Getenv("STRIPE_PRICE_PLUS"),
		StripePricePro:       os.Getenv("STRIPE_PRICE_PRO"),
		KafkaEnabled:         getEnvAsBool("KAFKA_ENABLED", true),
		KafkaBootstrapServer: getEnv("KAFKA_BOOTSTRAP_SERVERS", getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaBrokers:         splitCSV(getEnv("KAFKA_BROKERS", getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"))),
		KafkaClientID:        getEnv("KAFKA_CLIENT_ID", "subscriptions-service"),
		KafkaUsername:        os.Getenv("KAFKA_USERNAME"),
		KafkaPassword:        os.Getenv("KAFKA_PASSWORD"),
		KafkaSecurityProto:   getEnv("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT"),
		KafkaSASLMechanism:   getEnv("KAFKA_SASL_MECHANISM", ""),
		KafkaCACert:          os.Getenv("KAFKA_CA_CERT"),
		KafkaCACertPath:      os.Getenv("KAFKA_CA_CERT_PATH"),
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

func getServerPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return getEnv("SERVER_PORT", "8081")
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

func getEnvAsBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return defaultValue
	}

	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}
