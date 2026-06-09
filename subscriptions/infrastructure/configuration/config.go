package configuration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	ServiceName                       string
	ConfigServiceURL                  string
	ConfigServiceTimeout              time.Duration
	ServerPort                        string
	AllowedOrigins                    []string
	DatabaseURL                       string
	StripePublishableKey              string
	StripeSecretKey                   string
	StripeWebhookSecret               string
	StripeCurrency                    string
	StripePriceFree                   string
	StripePricePlus                   string
	StripePricePro                    string
	KafkaEnabled                      bool
	KafkaBootstrapServer              string
	KafkaBrokers                      []string
	KafkaClientID                     string
	KafkaConsumerGroup                string
	KafkaUsername                     string
	KafkaPassword                     string
	KafkaSecurityProto                string
	KafkaSASLMechanism                string
	KafkaCACert                       string
	KafkaCACertPath                   string
	KafkaTopicSubscriptionCreated     string
	KafkaTopicSubscriptionCancelled   string
	KafkaTopicSubscriptionPlanChanged string
	KafkaTopicSubscriptionExpired     string
	KafkaTopicSubscriptionUpdated     string
}

func Load() AppConfig {
	_ = godotenv.Load()
	cfg := AppConfig{
		ServiceName:          getEnv("SERVICE_NAME", "subscriptions-service"),
		ConfigServiceURL:     strings.TrimRight(strings.TrimSpace(os.Getenv("CONFIG_SERVICE_URL")), "/"),
		ConfigServiceTimeout: getEnvAsDurationMS("CONFIG_SERVICE_TIMEOUT_MS", 3000),
	}

	remote := fetchRemoteConfig(cfg.ConfigServiceURL, cfg.ServiceName, cfg.ConfigServiceTimeout)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = buildPostgresURLFromEnv()
	}
	allowedOrigins := firstNonEmpty(
		os.Getenv("CORS_ALLOWED_ORIGINS"),
		os.Getenv("ALLOWED_ORIGINS"),
		remote.AllowedOrigins,
		"http://localhost:3000,http://localhost:5173",
	)
	kafkaBrokers := firstNonEmpty(
		os.Getenv("KAFKA_BROKERS"),
		os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		strings.Join(remote.KafkaBrokers, ","),
		remote.KafkaBootstrapServer,
		"localhost:9092",
	)

	cfg.ServerPort = firstNonEmpty(os.Getenv("PORT"), os.Getenv("SERVER_PORT"), remote.ServerPort, "8080")
	cfg.AllowedOrigins = splitCSV(allowedOrigins)
	cfg.DatabaseURL = firstNonEmpty(databaseURL, remote.DatabaseURL)
	cfg.StripePublishableKey = firstNonEmpty(os.Getenv("STRIPE_PUBLISHABLE_KEY"), remote.StripePublishableKey)
	cfg.StripeSecretKey = firstNonEmpty(os.Getenv("STRIPE_SECRET_KEY"), remote.StripeSecretKey)
	cfg.StripeWebhookSecret = firstNonEmpty(os.Getenv("STRIPE_WEBHOOK_SECRET"), remote.StripeWebhookSecret)
	cfg.StripeCurrency = firstNonEmpty(os.Getenv("STRIPE_CURRENCY"), remote.StripeCurrency, "PEN")
	cfg.StripePriceFree = firstNonEmpty(os.Getenv("STRIPE_PRICE_FREE"), remote.StripePriceFree)
	cfg.StripePricePlus = firstNonEmpty(os.Getenv("STRIPE_PRICE_PLUS"), remote.StripePricePlus)
	cfg.StripePricePro = firstNonEmpty(os.Getenv("STRIPE_PRICE_PRO"), remote.StripePricePro)
	cfg.KafkaEnabled = getEnvAsBoolWithFallback("KAFKA_ENABLED", remote.KafkaEnabled, true)
	cfg.KafkaBootstrapServer = kafkaBrokers
	cfg.KafkaBrokers = splitCSV(kafkaBrokers)
	cfg.KafkaClientID = firstNonEmpty(os.Getenv("KAFKA_CLIENT_ID"), remote.KafkaClientID, "subscriptions-service")
	cfg.KafkaConsumerGroup = firstNonEmpty(os.Getenv("KAFKA_CONSUMER_GROUP"), remote.KafkaConsumerGroup, "subscriptions-service")
	cfg.KafkaUsername = firstNonEmpty(os.Getenv("KAFKA_USERNAME"), remote.KafkaUsername)
	cfg.KafkaPassword = firstNonEmpty(os.Getenv("KAFKA_PASSWORD"), remote.KafkaPassword)
	cfg.KafkaSecurityProto = firstNonEmpty(os.Getenv("KAFKA_SECURITY_PROTOCOL"), remote.KafkaSecurityProto, "PLAINTEXT")
	cfg.KafkaSASLMechanism = firstNonEmpty(os.Getenv("KAFKA_SASL_MECHANISM"), remote.KafkaSASLMechanism)
	cfg.KafkaCACert = firstNonEmpty(os.Getenv("KAFKA_CA_CERT"), remote.KafkaCACert)
	cfg.KafkaCACertPath = firstNonEmpty(os.Getenv("KAFKA_CA_CERT_PATH"), remote.KafkaCACertPath)
	cfg.KafkaTopicSubscriptionCreated = firstNonEmpty(
		os.Getenv("KAFKA_TOPIC_SUBSCRIPTION_CREATED"),
		remote.KafkaTopicSubscriptionCreated,
		"subscription.created",
	)
	cfg.KafkaTopicSubscriptionCancelled = firstNonEmpty(
		os.Getenv("KAFKA_TOPIC_SUBSCRIPTION_CANCELLED"),
		remote.KafkaTopicSubscriptionCancelled,
		"subscription.cancelled",
	)
	cfg.KafkaTopicSubscriptionPlanChanged = firstNonEmpty(
		os.Getenv("KAFKA_TOPIC_SUBSCRIPTION_PLAN_CHANGED"),
		remote.KafkaTopicSubscriptionPlanChanged,
		"subscription.plan.changed",
	)
	cfg.KafkaTopicSubscriptionExpired = firstNonEmpty(
		os.Getenv("KAFKA_TOPIC_SUBSCRIPTION_EXPIRED"),
		remote.KafkaTopicSubscriptionExpired,
		"subscription.expired",
	)
	cfg.KafkaTopicSubscriptionUpdated = firstNonEmpty(
		os.Getenv("KAFKA_TOPIC_SUBSCRIPTION_UPDATED"),
		remote.KafkaTopicSubscriptionUpdated,
		"subscription.updated",
	)

	return cfg
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
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

func getEnvAsBoolWithFallback(key string, fallback *bool, defaultValue bool) bool {
	if strings.TrimSpace(os.Getenv(key)) == "" {
		if fallback != nil {
			return *fallback
		}
		return defaultValue
	}
	return getEnvAsBool(key, defaultValue)
}

func getEnvAsDurationMS(key string, defaultMS int) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return time.Duration(defaultMS) * time.Millisecond
	}
	ms, err := strconv.Atoi(value)
	if err != nil || ms <= 0 {
		return time.Duration(defaultMS) * time.Millisecond
	}
	return time.Duration(ms) * time.Millisecond
}

type remoteConfig struct {
	ServerPort                        string   `json:"serverPort"`
	AllowedOrigins                    string   `json:"allowedOrigins"`
	DatabaseURL                       string   `json:"databaseUrl"`
	StripePublishableKey              string   `json:"stripePublishableKey"`
	StripeSecretKey                   string   `json:"stripeSecretKey"`
	StripeWebhookSecret               string   `json:"stripeWebhookSecret"`
	StripeCurrency                    string   `json:"stripeCurrency"`
	StripePriceFree                   string   `json:"stripePriceFree"`
	StripePricePlus                   string   `json:"stripePricePlus"`
	StripePricePro                    string   `json:"stripePricePro"`
	KafkaEnabled                      *bool    `json:"kafkaEnabled"`
	KafkaBootstrapServer              string   `json:"kafkaBootstrapServers"`
	KafkaBrokers                      []string `json:"kafkaBrokers"`
	KafkaClientID                     string   `json:"kafkaClientId"`
	KafkaConsumerGroup                string   `json:"kafkaConsumerGroup"`
	KafkaUsername                     string   `json:"kafkaUsername"`
	KafkaPassword                     string   `json:"kafkaPassword"`
	KafkaSecurityProto                string   `json:"kafkaSecurityProtocol"`
	KafkaSASLMechanism                string   `json:"kafkaSaslMechanism"`
	KafkaCACert                       string   `json:"kafkaCaCert"`
	KafkaCACertPath                   string   `json:"kafkaCaCertPath"`
	KafkaTopicSubscriptionCreated     string   `json:"topicSubscriptionCreated"`
	KafkaTopicSubscriptionCancelled   string   `json:"topicSubscriptionCancelled"`
	KafkaTopicSubscriptionPlanChanged string   `json:"topicSubscriptionPlanChanged"`
	KafkaTopicSubscriptionExpired     string   `json:"topicSubscriptionExpired"`
	KafkaTopicSubscriptionUpdated     string   `json:"topicSubscriptionUpdated"`
}

func fetchRemoteConfig(configServiceURL, serviceName string, timeout time.Duration) remoteConfig {
	if configServiceURL == "" || serviceName == "" {
		return remoteConfig{}
	}
	client := &http.Client{Timeout: timeout}
	serviceCfg := getConfigEndpoint(client, configServiceURL+"/api/v1/config/"+serviceName)
	servicesCatalogCfg := getServiceFromCatalog(client, configServiceURL+"/api/v1/config/services", serviceName)
	kafkaCfg := getConfigEndpoint(client, configServiceURL+"/api/v1/config/kafka")

	merged := kafkaCfg
	mergeRemoteConfig(&merged, servicesCatalogCfg)
	mergeRemoteConfig(&merged, serviceCfg)
	return merged
}

func getServiceFromCatalog(client *http.Client, endpoint, serviceName string) remoteConfig {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return remoteConfig{}
	}
	resp, err := client.Do(req)
	if err != nil {
		return remoteConfig{}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return remoteConfig{}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return remoteConfig{}
	}

	var catalog []map[string]json.RawMessage
	if err = json.Unmarshal(body, &catalog); err != nil {
		var wrapped map[string]json.RawMessage
		if unwrapErr := json.Unmarshal(body, &wrapped); unwrapErr != nil {
			return remoteConfig{}
		}
		for _, key := range []string{"data", "services"} {
			raw, ok := wrapped[key]
			if !ok {
				continue
			}
			if unmarshalErr := json.Unmarshal(raw, &catalog); unmarshalErr == nil {
				break
			}
		}
	}
	for _, item := range catalog {
		if matchesServiceName(item, serviceName) {
			bytes, marshalErr := json.Marshal(item)
			if marshalErr != nil {
				return remoteConfig{}
			}
			return parseRemoteConfigBody(bytes)
		}
	}
	return remoteConfig{}
}

func matchesServiceName(item map[string]json.RawMessage, serviceName string) bool {
	for _, key := range []string{"name", "service", "serviceName"} {
		raw, ok := item[key]
		if !ok {
			continue
		}
		var value string
		if err := json.Unmarshal(raw, &value); err == nil && strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(serviceName)) {
			return true
		}
	}
	return false
}

func getConfigEndpoint(client *http.Client, endpoint string) remoteConfig {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return remoteConfig{}
	}
	resp, err := client.Do(req)
	if err != nil {
		return remoteConfig{}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return remoteConfig{}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return remoteConfig{}
	}
	return parseRemoteConfigBody(body)
}

func parseRemoteConfigBody(body []byte) remoteConfig {
	var direct remoteConfig
	if err := json.Unmarshal(body, &direct); err == nil && !isRemoteConfigEmpty(direct) {
		return direct
	}

	var wrapped map[string]json.RawMessage
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return remoteConfig{}
	}
	for _, key := range []string{"data", "config", "service", "kafka"} {
		raw, ok := wrapped[key]
		if !ok {
			continue
		}
		var nested remoteConfig
		if err := json.Unmarshal(raw, &nested); err == nil {
			return nested
		}
	}
	return remoteConfig{}
}

func mergeRemoteConfig(target *remoteConfig, source remoteConfig) {
	target.ServerPort = firstNonEmpty(source.ServerPort, target.ServerPort)
	target.AllowedOrigins = firstNonEmpty(source.AllowedOrigins, target.AllowedOrigins)
	target.DatabaseURL = firstNonEmpty(source.DatabaseURL, target.DatabaseURL)
	target.StripePublishableKey = firstNonEmpty(source.StripePublishableKey, target.StripePublishableKey)
	target.StripeSecretKey = firstNonEmpty(source.StripeSecretKey, target.StripeSecretKey)
	target.StripeWebhookSecret = firstNonEmpty(source.StripeWebhookSecret, target.StripeWebhookSecret)
	target.StripeCurrency = firstNonEmpty(source.StripeCurrency, target.StripeCurrency)
	target.StripePriceFree = firstNonEmpty(source.StripePriceFree, target.StripePriceFree)
	target.StripePricePlus = firstNonEmpty(source.StripePricePlus, target.StripePricePlus)
	target.StripePricePro = firstNonEmpty(source.StripePricePro, target.StripePricePro)
	target.KafkaBootstrapServer = firstNonEmpty(source.KafkaBootstrapServer, target.KafkaBootstrapServer)
	if len(source.KafkaBrokers) > 0 {
		target.KafkaBrokers = source.KafkaBrokers
	}
	target.KafkaClientID = firstNonEmpty(source.KafkaClientID, target.KafkaClientID)
	target.KafkaConsumerGroup = firstNonEmpty(source.KafkaConsumerGroup, target.KafkaConsumerGroup)
	target.KafkaUsername = firstNonEmpty(source.KafkaUsername, target.KafkaUsername)
	target.KafkaPassword = firstNonEmpty(source.KafkaPassword, target.KafkaPassword)
	target.KafkaSecurityProto = firstNonEmpty(source.KafkaSecurityProto, target.KafkaSecurityProto)
	target.KafkaSASLMechanism = firstNonEmpty(source.KafkaSASLMechanism, target.KafkaSASLMechanism)
	target.KafkaCACert = firstNonEmpty(source.KafkaCACert, target.KafkaCACert)
	target.KafkaCACertPath = firstNonEmpty(source.KafkaCACertPath, target.KafkaCACertPath)
	target.KafkaTopicSubscriptionCreated = firstNonEmpty(source.KafkaTopicSubscriptionCreated, target.KafkaTopicSubscriptionCreated)
	target.KafkaTopicSubscriptionCancelled = firstNonEmpty(source.KafkaTopicSubscriptionCancelled, target.KafkaTopicSubscriptionCancelled)
	target.KafkaTopicSubscriptionPlanChanged = firstNonEmpty(source.KafkaTopicSubscriptionPlanChanged, target.KafkaTopicSubscriptionPlanChanged)
	target.KafkaTopicSubscriptionExpired = firstNonEmpty(source.KafkaTopicSubscriptionExpired, target.KafkaTopicSubscriptionExpired)
	target.KafkaTopicSubscriptionUpdated = firstNonEmpty(source.KafkaTopicSubscriptionUpdated, target.KafkaTopicSubscriptionUpdated)
	if source.KafkaEnabled != nil {
		target.KafkaEnabled = source.KafkaEnabled
	}
}

func isRemoteConfigEmpty(cfg remoteConfig) bool {
	return strings.TrimSpace(cfg.ServerPort) == "" &&
		strings.TrimSpace(cfg.AllowedOrigins) == "" &&
		strings.TrimSpace(cfg.DatabaseURL) == "" &&
		strings.TrimSpace(cfg.StripePublishableKey) == "" &&
		strings.TrimSpace(cfg.StripeSecretKey) == "" &&
		strings.TrimSpace(cfg.StripeWebhookSecret) == "" &&
		strings.TrimSpace(cfg.StripeCurrency) == "" &&
		strings.TrimSpace(cfg.StripePriceFree) == "" &&
		strings.TrimSpace(cfg.StripePricePlus) == "" &&
		strings.TrimSpace(cfg.StripePricePro) == "" &&
		strings.TrimSpace(cfg.KafkaBootstrapServer) == "" &&
		len(cfg.KafkaBrokers) == 0 &&
		strings.TrimSpace(cfg.KafkaClientID) == "" &&
		strings.TrimSpace(cfg.KafkaConsumerGroup) == "" &&
		strings.TrimSpace(cfg.KafkaUsername) == "" &&
		strings.TrimSpace(cfg.KafkaPassword) == "" &&
		strings.TrimSpace(cfg.KafkaSecurityProto) == "" &&
		strings.TrimSpace(cfg.KafkaSASLMechanism) == "" &&
		strings.TrimSpace(cfg.KafkaCACert) == "" &&
		strings.TrimSpace(cfg.KafkaCACertPath) == "" &&
		strings.TrimSpace(cfg.KafkaTopicSubscriptionCreated) == "" &&
		strings.TrimSpace(cfg.KafkaTopicSubscriptionCancelled) == "" &&
		strings.TrimSpace(cfg.KafkaTopicSubscriptionPlanChanged) == "" &&
		strings.TrimSpace(cfg.KafkaTopicSubscriptionExpired) == "" &&
		strings.TrimSpace(cfg.KafkaTopicSubscriptionUpdated) == "" &&
		cfg.KafkaEnabled == nil
}
