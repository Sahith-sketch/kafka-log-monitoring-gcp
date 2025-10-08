package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"audit-logs-monitoring/internal/config"
	"audit-logs-monitoring/internal/handlers"
	"audit-logs-monitoring/internal/services"
	"audit-logs-monitoring/pkg/logger"
)

func main() {
	
	cfg := config.Load()
	
	// Initialize services
	var kafkaProducer services.KafkaPublisher
	if cfg.IsKafkaEnabled {
		kp, err := services.NewKafkaProducer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.TLSCACertSecret, cfg.TLSClientCertSecret, cfg.TLSClientKeySecret, cfg.ProjectID)
		if err != nil {
			logger.WithError(err).Error("Failed to create Kafka producer")
			os.Exit(1)
		}
		defer kp.Close()
		kafkaProducer = kp
		logger.Info("Kafka producer enabled")
	} else {
		kafkaProducer = nil
		logger.Info("Kafka producer disabled")
	}

	var gcsStorage services.StorageService
	if cfg.IsGCSEnabled {
		gcs, err := services.NewGCSStorage(cfg.GCSBucket, cfg.BucketPath)
		if err != nil {
			logger.WithError(err).Error("Failed to create GCS storage")
			os.Exit(1)
		}
		defer gcs.Close()
		gcsStorage = gcs
		logger.Info("GCS storage enabled")
	} else {
		gcsStorage = nil
		logger.Info("GCS storage disabled")
	}
	
	detector := services.NewIntrusionDetector()
	var emailNotifier services.AlertNotifier
	if cfg.IsNotificationsEnabled {
		emailNotifier = services.NewEmailNotifier(cfg)
		logger.Info("Email notifications enabled")
	} else {
		emailNotifier = nil
		logger.Info("Email notifications disabled")
	}
	
	batchProcessor := services.NewBatchProcessor(
		cfg.BatchSize,
		cfg.BatchTime,
		cfg.WorkerCount,
		kafkaProducer,
		detector,
		emailNotifier,
		gcsStorage,
		cfg.IsKafkaEnabled,
		cfg.IsGCSEnabled,
		cfg.IsNotificationsEnabled,
	)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Start batch processor
	batchProcessor.Start(ctx)
	
	// Setup HTTP handlers
	pubsubHandler := handlers.NewPubSubHandler(batchProcessor)
	
	http.HandleFunc("/", pubsubHandler.HandlePubSub)
	http.HandleFunc("/health", pubsubHandler.HealthCheck)
	
	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	
	go func() {
		logger.WithField("port", cfg.Port).Info("Server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("Server failed to start")
			os.Exit(1)
		}
	}()
	
	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	logger.Info("Shutting down...")
	cancel()
	batchProcessor.Stop()
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Server shutdown error")
	}
	
	logger.Info("Server stopped")
}
