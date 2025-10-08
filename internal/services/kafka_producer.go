package services

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"time"
	"audit-logs-monitoring/internal/models"
	"audit-logs-monitoring/pkg/logger"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string, caCertSecret, clientCertSecret, clientKeySecret, projectID string) (*KafkaProducer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
	}

	// Configure TLS if certificate secrets are provided
	if caCertSecret != "" && clientCertSecret != "" && clientKeySecret != "" && projectID != "" {
		tlsConfig, err := createTLSConfigFromSecrets(caCertSecret, clientCertSecret, clientKeySecret, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		writer.Transport = &kafka.Transport{
			TLS: tlsConfig,
		}
	}

	return &KafkaProducer{
		writer: writer,
	}, nil
}

func createTLSConfigFromSecrets(caCertSecret, clientCertSecret, clientKeySecret, projectID string) (*tls.Config, error) {
	sm, err := NewSecretManager(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager: %w", err)
	}
	defer sm.Close()

	caCert, err := sm.GetSecret(caCertSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get CA cert: %w", err)
	}

	clientCert, err := sm.GetSecret(clientCertSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get client cert: %w", err)
	}

	clientKey, err := sm.GetSecret(clientKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get client key: %w", err)
	}

	return createTLSConfig(caCert, clientCert, clientKey)
}

func createTLSConfig(caCert, clientCert, clientKey string) (*tls.Config, error) {
	cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM([]byte(caCert)) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}, nil
}

func (kp *KafkaProducer) PublishBatch(logs []models.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(logs))
	var marshalErrors []error
	
	for _, auditLog := range logs {
		data, err := json.Marshal(auditLog)
		if err != nil {
			logger.WithError(err).Error("Failed to marshal audit log")
			marshalErrors = append(marshalErrors, err)
			continue
		}

		messages = append(messages, kafka.Message{
			Value: data,
		})
	}

	if len(messages) == 0 {
		return fmt.Errorf("no valid messages to publish, %d marshal errors", len(marshalErrors))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := kp.writer.WriteMessages(ctx, messages...); err != nil {
		logger.WithError(err).WithField("message_count", len(messages)).Error("Failed to publish messages to Kafka")
		return fmt.Errorf("failed to publish %d messages: %w", len(messages), err)
	}

	logger.WithField("message_count", len(messages)).Info("Successfully published messages to Kafka")
	return nil
}

func (kp *KafkaProducer) Close() error {
	if err := kp.writer.Close(); err != nil {
		logger.WithError(err).Error("Failed to close Kafka writer")
		return fmt.Errorf("failed to close Kafka writer: %w", err)
	}
	logger.Info("Kafka writer closed successfully")
	return nil
}