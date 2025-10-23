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
		RequiredAcks: kafka.RequireOne,
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



func (kp *KafkaProducer) PublishKafkaMessage(msg models.KafkaMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal kafka message")
		return fmt.Errorf("failed to marshal kafka message: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(msg.InsertId),
		Value: data,
		Time:  time.Now(),
	}

	// Retry with exponential backoff
	for attempt := 1; attempt <= 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := kp.writer.WriteMessages(ctx, message)
		cancel()
		
		if err == nil {
			logger.WithField("insertId", msg.InsertId).Info("Successfully published message to Kafka")
			return nil
		}
		
		logger.WithError(err).WithField("insertId", msg.InsertId).WithField("attempt", attempt).Error("Failed to publish message to Kafka")
		
		if attempt < 3 {
			backoff := time.Duration(attempt*attempt) * time.Second // 1s, 4s
			logger.WithField("backoff", backoff).Info("Retrying Kafka publish")
			time.Sleep(backoff)
		}
	}
	
	return fmt.Errorf("failed to publish message after 3 attempts")
}

func (kp *KafkaProducer) PublishLog(log models.KafkaMessage) error {
	data, err := json.Marshal(log)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal audit log")
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(log.InsertId),
		Value: data,
		Time:  time.Now(),
	}

	// Retry with exponential backoff
	for attempt := 1; attempt <= 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := kp.writer.WriteMessages(ctx, message)
		cancel()
		
		if err == nil {
			logger.WithField("insertId", log.InsertId).Info("Successfully published message to Kafka")
			return nil
		}
		
		logger.WithError(err).WithField("insertId", log.InsertId).WithField("attempt", attempt).Error("Failed to publish message to Kafka")
		
		if attempt < 3 {
			backoff := time.Duration(attempt*attempt) * time.Second // 1s, 4s
			logger.WithField("backoff", backoff).Info("Retrying Kafka publish")
			time.Sleep(backoff)
		}
	}
	
	return fmt.Errorf("failed to publish message after 3 attempts")
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

func (kp *KafkaProducer) Close() error {
	if err := kp.writer.Close(); err != nil {
		logger.WithError(err).Error("Failed to close Kafka writer")
		return fmt.Errorf("failed to close Kafka writer: %w", err)
	}
	logger.Info("Kafka writer closed successfully")
	return nil
}