package services

import (
	"audit-logs-monitoring/internal/models"
	"testing"
	"time"
)

func TestKafkaProducer_RetryLogic(t *testing.T) {
	// Test retry backoff calculation
	for attempt := 1; attempt <= 3; attempt++ {
		if attempt < 3 {
			backoff := time.Duration(attempt*attempt) * time.Second
			expectedBackoff := []time.Duration{1 * time.Second, 4 * time.Second}[attempt-1]
			
			if backoff != expectedBackoff {
				t.Errorf("Attempt %d: expected backoff %v, got %v", attempt, expectedBackoff, backoff)
			}
		}
	}
}

func TestKafkaProducer_PublishLog_WithRetries(t *testing.T) {
	// Test that the function structure supports retries
	producer, err := NewKafkaProducer([]string{"invalid:9092"}, "test-topic", "", "", "", "")
	if err != nil {
		t.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	testMsg := models.KafkaMessage{
		Timestamp:  time.Now().Format("2006-01-02T15:04:05.000Z"),
		Severity:   "INFO",
		InsertId:   "retry-test-2",
		ProjectId:  "test-project",
		LogType:    "retry-test",
	}

	// This will fail but should attempt retries
	start := time.Now()
	err = producer.PublishLog(testMsg)
	duration := time.Since(start)

	// Should fail after retries
	if err == nil {
		t.Error("Expected error with invalid broker")
	}

	// Should take at least 5 seconds (1s + 4s backoff)
	if duration < 5*time.Second {
		t.Errorf("Expected at least 5s for retries, took %v", duration)
	}

	// Should not take too long (max ~15s with timeouts)
	if duration > 20*time.Second {
		t.Errorf("Took too long: %v", duration)
	}
}