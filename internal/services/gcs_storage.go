package services

import (
	"audit-logs-monitoring/internal/models"
	"context"
	"encoding/csv"
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

func (gcs *GCSStorage) SaveBatchAsCSV(logs []models.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}

	fileName := fmt.Sprintf("audit-logs-%s.csv", time.Now().Format("2006-01-02-15-04-05"))
	if gcs.bucketPath != "" {
		fileName = fmt.Sprintf("%s/%s", gcs.bucketPath, fileName)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	obj := gcs.client.Bucket(gcs.bucketName).Object(fileName)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write CSV header with all AuditLog fields
	header := []string{
		"Timestamp", "ReceiveTimestamp", "Severity", "LogName", "InsertId",
		"Resource", "ProtoPayload", "HttpRequest", "Operation", "Trace",
		"SpanId", "SourceLocation", "Labels", "TextPayload", "JsonPayload",
	}
	if err := csvWriter.Write(header); err != nil {
		logger.Infof("Failed to write CSV header: %v", err)
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows with all fields
	for _, log := range logs {
		row := []string{
			log.Timestamp.Format(time.RFC3339),
			log.ReceiveTimestamp.Format(time.RFC3339),
			log.Severity,
			log.LogName,
			log.InsertId,
			toJSON(log.Resource),
			toJSON(log.ProtoPayload),
			toJSON(log.HttpRequest),
			toJSON(log.Operation),
			log.Trace,
			log.SpanId,
			toJSON(log.SourceLocation),
			toJSON(log.Labels),
			log.TextPayload,
			toJSON(log.JsonPayload),
		}
		if err := csvWriter.Write(row); err != nil {
			logger.Infof("Failed to write CSV row: %v", err)
			continue
		}
	}

	if err := writer.Close(); err != nil {
		logger.Infof("Failed to close GCS writer for file %s: %v", fileName, err)
		return fmt.Errorf("failed to save CSV to GCS: %w", err)
	}

	logger.Infof("Successfully saved batch to GCS: file=%s, records=%d, bucket=%s", fileName, len(logs), gcs.bucketName)

	return nil
}

func toJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

func (gcs *GCSStorage) Close() error {
	if err := gcs.client.Close(); err != nil {
		logger.Infof("Failed to close GCS client: %v", err)
		return fmt.Errorf("failed to close GCS client: %w", err)
	}
	logger.Info("GCS client closed successfully")
	return nil
}
