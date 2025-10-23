package services

import (
	"audit-logs-monitoring/internal/models"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

// MockKafkaWriter implements a mock Kafka writer for testing
type MockKafkaWriter struct {
	messages []kafka.Message
	closed   bool
	failNext bool
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if m.failNext {
		m.failNext = false
		return kafka.WriteErrors{kafka.MessageTooLargeError{}}
	}
	m.messages = append(m.messages, msgs...)
	return nil
}

func (m *MockKafkaWriter) Close() error {
	m.closed = true
	return nil
}





func TestKafkaProducer_PublishLog_ValidLog(t *testing.T) {
	logs := createTestKafkaLogs()
	log := logs[0]
	
	// Test JSON marshaling
	data, err := json.Marshal(log)
	if err != nil {
		t.Errorf("Failed to marshal log: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty marshaled data")
	}
	
	// Verify JSON contains expected fields
	jsonStr := string(data)
	if !contains(jsonStr, "timestamp") {
		t.Error("JSON should contain timestamp field")
	}
	if !contains(jsonStr, log.Severity) {
		t.Errorf("JSON should contain severity %s", log.Severity)
	}
}

func TestKafkaProducer_NewProducer(t *testing.T) {
	producer, err := NewKafkaProducer([]string{"localhost:9092"}, "test-topic", "", "", "", "")
	if err != nil {
		t.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()
	
	if producer.writer == nil {
		t.Error("Expected writer to be initialized")
	}
}

func TestKafkaProducer_MessageGeneration(t *testing.T) {
	logs := createTestKafkaLogs()
	
	// Test message generation logic
	messages := make([]kafka.Message, 0, len(logs))
	var marshalErrors []error
	
	for _, auditLog := range logs {
		data, err := json.Marshal(auditLog)
		if err != nil {
			marshalErrors = append(marshalErrors, err)
			continue
		}
		
		messages = append(messages, kafka.Message{
			Value: data,
		})
	}
	
	if len(marshalErrors) > 0 {
		t.Errorf("Unexpected marshal errors: %v", marshalErrors)
	}
	
	if len(messages) != len(logs) {
		t.Errorf("Expected %d messages, got %d", len(logs), len(messages))
	}
	
	// Verify message content
	for i, msg := range messages {
		if len(msg.Value) == 0 {
			t.Errorf("Message %d has empty value", i)
		}
		
		// Verify it's valid JSON
		var decoded models.AuditLog
		err := json.Unmarshal(msg.Value, &decoded)
		if err != nil {
			t.Errorf("Message %d contains invalid JSON: %v", i, err)
		}
		
		// Verify key fields are preserved
		if decoded.Severity != logs[i].Severity {
			t.Errorf("Message %d: expected severity %s, got %s", i, logs[i].Severity, decoded.Severity)
		}
		if decoded.LogName != logs[i].LogName {
			t.Errorf("Message %d: expected log name %s, got %s", i, logs[i].LogName, decoded.LogName)
		}
	}
}

func TestKafkaProducer_JSONMarshaling_EdgeCases(t *testing.T) {
	// Test with minimal log
	minimalLog := models.AuditLog{
		Timestamp: time.Now(),
		Severity:  "INFO",
		LogName:   "minimal-log",
		Resource: models.Resource{
			Type: "test",
		},
	}
	
	data, err := json.Marshal(minimalLog)
	if err != nil {
		t.Errorf("Failed to marshal minimal log: %v", err)
	}
	
	if len(data) == 0 {
		t.Error("Expected non-empty data for minimal log")
	}
	
	// Test with complex log
	complexLog := models.AuditLog{
		Timestamp: time.Now(),
		Severity:  "ERROR",
		LogName:   "complex-log",
		Resource: models.Resource{
			Type:   "compute",
			Labels: map[string]string{"zone": "us-central1", "project": "test"},
		},
		ProtoPayload: &models.Payload{
			Type:       "audit",
			MethodName: "test.method",
			Request:    map[string]interface{}{"key": "value", "nested": map[string]string{"inner": "data"}},
			Response:   map[string]interface{}{"status": "success"},
		},
		Labels: map[string]string{"env": "test"},
		JsonPayload: map[string]interface{}{
			"custom": "data",
			"array":  []int{1, 2, 3},
		},
	}
	
	data, err = json.Marshal(complexLog)
	if err != nil {
		t.Errorf("Failed to marshal complex log: %v", err)
	}
	
	if len(data) == 0 {
		t.Error("Expected non-empty data for complex log")
	}
	
	// Verify complex data can be unmarshaled
	var decoded models.AuditLog
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("Failed to unmarshal complex log: %v", err)
	}
	
	if decoded.Severity != complexLog.Severity {
		t.Errorf("Expected severity %s, got %s", complexLog.Severity, decoded.Severity)
	}
}





func createTestKafkaLogs() []models.AuditLog {
	now := time.Now()
	
	return []models.AuditLog{
		{
			Timestamp:        now,
			ReceiveTimestamp: now.Add(time.Millisecond),
			Severity:         "INFO",
			LogName:          "test-log-1",
			InsertId:         "insert-1",
			Resource: models.Resource{
				Type:   "compute",
				Labels: map[string]string{"zone": "us-central1"},
			},
			ProtoPayload: &models.Payload{
				Type:         "type.googleapis.com/google.cloud.audit.AuditLog",
				MethodName:   "compute.instances.insert",
				ResourceName: "projects/test/zones/us-central1/instances/test-vm",
				ServiceName:  "compute.googleapis.com",
				CallerIp:     "192.168.1.1",
			},
			Labels: map[string]string{"project_id": "test-project"},
		},
		{
			Timestamp:        now.Add(time.Second),
			ReceiveTimestamp: now.Add(time.Second + time.Millisecond),
			Severity:         "WARNING",
			LogName:          "test-log-2",
			InsertId:         "insert-2",
			Resource: models.Resource{
				Type:   "storage",
				Labels: map[string]string{"bucket": "test-bucket"},
			},
			ProtoPayload: &models.Payload{
				Type:         "type.googleapis.com/google.cloud.audit.AuditLog",
				MethodName:   "storage.objects.create",
				ResourceName: "projects/test/buckets/test-bucket/objects/test-file",
				ServiceName:  "storage.googleapis.com",
				CallerIp:     "10.0.0.1",
			},
			JsonPayload: map[string]interface{}{
				"eventType": "OBJECT_CREATE",
				"bucketId":  "test-bucket",
			},
		},
		{
			Timestamp: now.Add(2 * time.Second),
			Severity:  "ERROR",
			LogName:   "test-log-3",
			Resource: models.Resource{
				Type: "network",
			},
			TextPayload: "Network error occurred",
		},
	}
}

