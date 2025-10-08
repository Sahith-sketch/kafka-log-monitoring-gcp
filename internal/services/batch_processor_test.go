package services

import (
	"context"
	"testing"
	"time"
	"audit-logs-monitoring/internal/models"
)

type mockKafkaProducer struct {
	published []models.AuditLog
}

func (m *mockKafkaProducer) PublishBatch(logs []models.AuditLog) error {
	m.published = append(m.published, logs...)
	return nil
}

func (m *mockKafkaProducer) Close() error {
	return nil
}

type mockEmailNotifier struct {
	alerts []*models.IntrusionAlert
}

func (m *mockEmailNotifier) SendAlert(alert *models.IntrusionAlert) error {
	m.alerts = append(m.alerts, alert)
	return nil
}

type mockStorageService struct {
	savedBatches [][]models.AuditLog
}

func (m *mockStorageService) SaveBatchAsCSV(logs []models.AuditLog) error {
	m.savedBatches = append(m.savedBatches, logs)
	return nil
}

func (m *mockStorageService) Close() error {
	return nil
}

func TestNewBatchProcessor(t *testing.T) {
	mockKafka := &mockKafkaProducer{}
	detector := NewIntrusionDetector()
	mockEmail := &mockEmailNotifier{}
	
	processor := NewBatchProcessor(10, 5*time.Second, 2, mockKafka, detector, mockEmail, &mockStorageService{}, true, true, true)
	
	if processor.batchSize != 10 {
		t.Errorf("Expected batch size 10, got %d", processor.batchSize)
	}
	if processor.workerCount != 2 {
		t.Errorf("Expected worker count 2, got %d", processor.workerCount)
	}
}

func TestBatchProcessor_ProcessLog(t *testing.T) {
	mockKafka := &mockKafkaProducer{}
	detector := NewIntrusionDetector()
	mockEmail := &mockEmailNotifier{}
	
	processor := NewBatchProcessor(2, 5*time.Second, 1, mockKafka, detector, mockEmail, &mockStorageService{}, true, true, true)
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	processor.Start(ctx)
	
	// Process logs
	log1 := models.AuditLog{
		Timestamp: time.Now(),
		Severity:  "INFO",
		ProtoPayload: &models.Payload{
			MethodName: "compute.instances.insert",
		},
	}
	
	log2 := models.AuditLog{
		Timestamp: time.Now(),
		Severity:  "INFO",
		ProtoPayload: &models.Payload{
			MethodName: "compute.firewalls.insert", // This should trigger alert
		},
	}
	
	processor.ProcessLog(log1)
	processor.ProcessLog(log2)
	
	// Wait for processing
	time.Sleep(100 * time.Millisecond)
	cancel()
	processor.Stop()
	
	// Verify processing
	if len(mockKafka.published) < 2 {
		t.Errorf("Expected at least 2 published logs, got %d", len(mockKafka.published))
	}
	
	if len(mockEmail.alerts) < 1 {
		t.Errorf("Expected at least 1 alert, got %d", len(mockEmail.alerts))
	}
}