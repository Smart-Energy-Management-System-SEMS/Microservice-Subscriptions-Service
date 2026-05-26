package kafka

import "github.com/segmentio/kafka-go"

type Publisher struct {
	brokers  []string
	clientID string
}

func NewPublisher(brokers []string, clientID string) *Publisher {
	return &Publisher{brokers: brokers, clientID: clientID}
}

func (p *Publisher) Publish(topic string, payload []byte) error {
	if len(p.brokers) == 0 {
		return nil
	}
	writer := &kafka.Writer{Addr: kafka.TCP(p.brokers...), Topic: topic, Balancer: &kafka.LeastBytes{}}
	defer writer.Close()
	return writer.WriteMessages(nil, kafka.Message{Value: payload})
}
