package services

import (
	"audit-logs-monitoring/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"audit-logs-monitoring/pkg/logger"

	"cloud.google.com/go/storage"
)

type GCSStorage struct {
	client     *storage.Client
	bucketName string
	bucketPath string
}

func NewGCSStorage(bucketName, bucketPath string) (*GCSStorage, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSStorage{
		client:     client,
		bucketName: bucketName,
		bucketPath: bucketPath,
	}, nil
}

func (gcs *GCSStorage) SaveLog(log models.AuditLog) error {
	fileName := fmt.Sprintf("audit-log-%s-%s.json", log.InsertId, time.Now().Format("2006-01-02-15-04-05"))
	if gcs.bucketPath != "" {
		fileName = fmt.Sprintf("%s/%s", gcs.bucketPath, fileName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	obj := gcs.client.Bucket(gcs.bucketName).Object(fileName)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	data, err := json.Marshal(log)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal audit log")
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		logger.WithError(err).Error("Failed to write to GCS")
		return fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		logger.WithError(err).WithField("fileName", fileName).Error("Failed to close GCS writer")
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	logger.WithField("fileName", fileName).WithField("insertId", log.InsertId).Info("Successfully saved log to GCS")
	return nil
}



func (gcs *GCSStorage) Close() error {
	if err := gcs.client.Close(); err != nil {
		logger.WithError(err).Error("Failed to close GCS client")
		return fmt.Errorf("failed to close GCS client: %w", err)
	}
	logger.Info("GCS client closed successfully")
	return nil
}
