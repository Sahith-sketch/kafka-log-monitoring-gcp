package main

import (
	"audit-logs-monitoring/internal/handlers"
	"audit-logs-monitoring/internal/services"
	"audit-logs-monitoring/pkg/logger"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	isGCSEnabled := os.Getenv("GCS_ENABLED") == "true"
	isKafkaEnabled := os.Getenv("KAFKA_ENABLED") == "true"

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var gcsStorage *services.GCSStorage
	if isGCSEnabled {
		gcsBucket := os.Getenv("GCS_BUCKET")
		if gcsBucket == "" {
			gcsBucket = "audit-logs-bucket"
		}

		gcsPath := os.Getenv("GCS_PATH")
		if gcsPath == "" {
			gcsPath = "audit-logs"
		}

		var err error
		gcsStorage, err = services.NewGCSStorage(gcsBucket, gcsPath)
		if err != nil {
			log.Fatalf("Failed to initialize GCS storage: %v", err)
		}
		defer gcsStorage.Close()
		logger.Infof("GCS storage enabled - Bucket: %s, Path: %s", gcsBucket, gcsPath)
	} else {
		logger.Info("GCS storage disabled")
	}

	var kafkaProducer *services.KafkaProducer
	if isKafkaEnabled {
		kafkaBrokers := os.Getenv("KAFKA_BROKERS")
		if kafkaBrokers == "" {
			kafkaBrokers = "localhost:9092"
		}

		kafkaTopic := os.Getenv("KAFKA_TOPIC")
		if kafkaTopic == "" {
			kafkaTopic = "audit-logs"
		}

		caCertSecret := os.Getenv("KAFKA_CA_CERT_SECRET")
		clientCertSecret := os.Getenv("KAFKA_CLIENT_CERT_SECRET")
		clientKeySecret := os.Getenv("KAFKA_CLIENT_KEY_SECRET")
		projectID := os.Getenv("PROJECT_ID")

		brokerList := strings.Split(kafkaBrokers, ",")
		var err error
		kafkaProducer, err = services.NewKafkaProducer(brokerList, kafkaTopic, caCertSecret, clientCertSecret, clientKeySecret, projectID)
		if err != nil {
			log.Fatalf("Failed to initialize Kafka producer: %v", err)
		}
		defer kafkaProducer.Close()
		logger.Infof("Kafka enabled - Brokers: %s, Topic: %s", kafkaBrokers, kafkaTopic)
	} else {
		logger.Info("Kafka disabled")
	}

	processor := services.NewProcessor(gcsStorage, kafkaProducer)
	defer processor.Close()

	handler := handlers.NewPubSubHandler(processor)

	http.HandleFunc("/", handler.HandlePubSub)
	http.HandleFunc("/health", handler.HealthCheck)

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		logger.Infof("Starting server on port %s", port)
		logger.Infof("GCS Enabled: %v, Kafka Enabled: %v", isGCSEnabled, isKafkaEnabled)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Server shutdown error")
	}

	logger.Info("Server stopped gracefully")
}
