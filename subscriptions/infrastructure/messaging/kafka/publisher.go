package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type Publisher struct {
	brokers  []string
	clientID string
	enabled  bool
	writer   *kafka.Writer
	initErr  error
}

type Config struct {
	Enabled          bool
	Brokers          []string
	ClientID         string
	Username         string
	Password         string
	SecurityProtocol string
	SASLMechanism    string
	CACert           string
	CACertPath       string
}

func NewPublisher(cfg Config) *Publisher {
	pub := &Publisher{brokers: cfg.Brokers, clientID: cfg.ClientID, enabled: cfg.Enabled}
	if !cfg.Enabled || len(cfg.Brokers) == 0 {
		return pub
	}

	transport := &kafka.Transport{}
	if mechanism, err := buildSASLMechanism(cfg); err == nil {
		transport.SASL = mechanism
	} else {
		pub.initErr = err
		return pub
	}
	if tlsConfig, err := buildTLSConfig(cfg); err == nil && tlsConfig != nil {
		transport.TLS = tlsConfig
	} else if err != nil {
		pub.initErr = err
		return pub
	}

	pub.writer = &kafka.Writer{
		Addr:      kafka.TCP(cfg.Brokers...),
		Balancer:  &kafka.LeastBytes{},
		Transport: transport,
	}

	return pub
}

func (p *Publisher) Publish(topic string, payload []byte) error {
	if !p.enabled || len(p.brokers) == 0 || p.writer == nil {
		if p.initErr != nil {
			return p.initErr
		}
		return nil
	}
	if topic == "" {
		return fmt.Errorf("kafka topic is required")
	}

	p.writer.Topic = topic
	return p.writer.WriteMessages(context.Background(), kafka.Message{Value: payload})
}

func buildSASLMechanism(cfg Config) (sasl.Mechanism, error) {
	securityProtocol := strings.ToUpper(strings.TrimSpace(cfg.SecurityProtocol))
	if securityProtocol == "" || securityProtocol == "PLAINTEXT" {
		return nil, nil
	}
	if cfg.Username == "" || cfg.Password == "" {
		return nil, fmt.Errorf("kafka username and password are required for %s", securityProtocol)
	}

	mechanism := strings.ToUpper(strings.TrimSpace(cfg.SASLMechanism))
	switch mechanism {
	case "", "PLAIN":
		return plain.Mechanism{Username: cfg.Username, Password: cfg.Password}, nil
	case "SCRAM-SHA-256":
		return scram.Mechanism(scram.SHA256, cfg.Username, cfg.Password)
	case "SCRAM-SHA-512":
		return scram.Mechanism(scram.SHA512, cfg.Username, cfg.Password)
	default:
		return nil, fmt.Errorf("unsupported kafka SASL mechanism: %s", mechanism)
	}
}

func buildTLSConfig(cfg Config) (*tls.Config, error) {
	securityProtocol := strings.ToUpper(strings.TrimSpace(cfg.SecurityProtocol))
	if securityProtocol != "SSL" && securityProtocol != "SASL_SSL" {
		return nil, nil
	}

	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}

	pemData := strings.TrimSpace(cfg.CACert)
	if pemData == "" && strings.TrimSpace(cfg.CACertPath) != "" {
		bytes, readErr := os.ReadFile(cfg.CACertPath)
		if readErr != nil {
			return nil, readErr
		}
		pemData = strings.TrimSpace(string(bytes))
	}
	if pemData != "" {
		pemData = strings.ReplaceAll(pemData, `\n`, "\n")
		if ok := pool.AppendCertsFromPEM([]byte(pemData)); !ok {
			return nil, fmt.Errorf("failed to parse KAFKA_CA_CERT")
		}
	}

	return &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: pool}, nil
}
