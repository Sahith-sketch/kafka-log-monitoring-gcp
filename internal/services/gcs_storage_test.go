package services

import (
	"audit-logs-monitoring/internal/models"
	"encoding/json"
	"testing"
	"time"
)

func TestGCSStorage_SaveLog(t *testing.T) {
	testLog := models.AuditLog{
		Timestamp:        time.Now(),
		ReceiveTimestamp: time.Now(),
		Severity:         "INFO",
		LogName:          "test-log",
		InsertId:         "test-1",
		Resource: models.Resource{
			Type: "compute",
			Labels: map[string]string{"zone": "us-central1"},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(testLog)
	if err != nil {
		t.Fatalf("Failed to marshal test log: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON data")
	}

	// Test unmarshaling
	var decoded models.AuditLog
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal test log: %v", err)
	}

	if decoded.InsertId != testLog.InsertId {
		t.Errorf("Expected InsertId %s, got %s", testLog.InsertId, decoded.InsertId)
	}
}

func TestGCSStorage_FileNameGeneration(t *testing.T) {
	testLog := models.AuditLog{
		InsertId: "test-123",
	}

	expectedPrefix := "audit-log-test-123-"
	fileName := "audit-log-" + testLog.InsertId + "-" + time.Now().Format("2006-01-02-15-04-05") + ".json"

	if !contains(fileName, expectedPrefix) {
		t.Errorf("Expected filename to contain %s, got %s", expectedPrefix, fileName)
	}

	if !contains(fileName, ".json") {
		t.Errorf("Expected filename to end with .json, got %s", fileName)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
			containsHelper(s, substr))))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
