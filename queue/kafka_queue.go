package queue

import (
	"context"
	"crypto/tls"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"time"
)

type KafkaQueue struct {
	ctx    context.Context
	writer *kafka.Writer
	reader *kafka.Reader
}

type KafkaConfig struct {
	// the kafka brokers to connect to
	KafkaURL []string
	// the kafka topic to write to and read from
	Topic string
	// the kafka.Transport that will be used to connect to kafka
	Transport *kafka.Transport
	// the kafka client id, defaults to "scrapper"
	ClientId string `default:"scrapper"`
	// the kafka group id, defaults to "scrapper"
	GroupID string `default:"scrapper"`
}

// NewKafka returns a new scrapper instance
func NewKafka(ctx context.Context, config KafkaConfig) (*KafkaQueue, error) {
	if config.ClientId == "" {
		config.ClientId = "scrapper"
	}
	if config.GroupID == "" {
		config.GroupID = "scrapper"
	}
	writer := kafka.Writer{
		Addr:      kafka.TCP(config.KafkaURL...),
		Topic:     config.Topic,
		Transport: config.Transport,
	}
	var saslMechanism sasl.Mechanism
	if config.Transport != nil {
		saslMechanism = config.Transport.SASL
	}
	var tlsConfig *tls.Config
	if config.Transport != nil {
		tlsConfig = config.Transport.TLS
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.KafkaURL,
		Topic:   config.Topic,
		GroupID: config.GroupID,
		Dialer: &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: saslMechanism,
			TLS:           tlsConfig,
			ClientID:      config.ClientId,
		},
	})
	return &KafkaQueue{
		ctx:    ctx,
		writer: &writer,
		reader: reader,
	}, nil
}

// Close closes the underlying kafka writer and reader.
func (s *KafkaQueue) Close() error {
	if err := s.writer.Close(); err != nil {
		return err
	}
	if err := s.reader.Close(); err != nil {
		return err
	}
	return nil
}

// WriteMessage writes a message to the underlying kafka writer.
// If a callback is provided, it will be called with the message value.
// If the callback returns an error, the message will not be written.
func (s *KafkaQueue) WriteMessage(msg []byte, callback func([]byte) error) error {
	err := s.writer.WriteMessages(s.ctx, kafka.Message{
		Value: msg,
	})
	if err != nil {
		return err
	}
	if callback == nil {
		return nil
	}
	err = callback(msg)
	if err != nil {
		return err
	}
	return nil
}

// ReadMessage reads a message from the underlying kafka reader.
// If a callback is provided, it will be called with the message value.
// If the callback returns an error, the message will not be committed.
func (s *KafkaQueue) ReadMessage(callback func([]byte) error) ([]byte, error) {
	m, err := s.reader.FetchMessage(s.ctx)
	if err != nil {
		return nil, err
	}
	if callback != nil {
		err = callback(m.Value)
		if err != nil {
			return nil, err
		}
	}
	if err = s.reader.CommitMessages(s.ctx, m); err != nil {
		return nil, err
	}
	return m.Value, nil
}
