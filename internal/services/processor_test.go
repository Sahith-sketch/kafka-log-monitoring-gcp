package services

import (
	"audit-logs-monitoring/internal/models"
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestProcessor_ProcessLog_BothEnabled(t *testing.T) {
	processor := NewProcessor(nil, nil)
	defer processor.Close()

	testLog := models.AuditLog{
		Timestamp:        time.Now(),
		ReceiveTimestamp: time.Now(),
		Severity:         "INFO",
		LogName:          "test-log",
		InsertId:         "processor-test-1",
		Resource: models.Resource{
			Type: "gce_instance",
			Labels: map[string]string{
				"instance_id": "123456789",
				"project_id":  "test-project",
			},
		},
		ProtoPayload: &models.Payload{
			MethodName:   "compute.instances.insert",
			ResourceName: "test-vm",
			CallerIp:     "192.168.1.1",
		},
	}

	ctx := context.Background()
	err := processor.ProcessLog(ctx, testLog)
	if err != nil {
		t.Errorf("Expected no error with nil services, got: %v", err)
	}
}

func TestProcessor_ProcessLog_ContextCancelled(t *testing.T) {
	processor := NewProcessor(nil, nil)
	defer processor.Close()

	testLog := models.AuditLog{
		InsertId: "cancelled-test",
		Severity: "INFO",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := processor.ProcessLog(ctx, testLog)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestProcessor_JSONSerialization(t *testing.T) {
	testLog := models.AuditLog{
		Timestamp:        time.Now(),
		ReceiveTimestamp: time.Now(),
		Severity:         "WARNING",
		LogName:          "projects/test/logs/audit",
		InsertId:         "json-test-1",
		Resource: models.Resource{
			Type: "gcs_bucket",
			Labels: map[string]string{
				"bucket_name": "test-bucket",
				"location":    "us-central1",
			},
		},
		ProtoPayload: &models.Payload{
			Type:         "audit.log",
			MethodName:   "storage.objects.create",
			ResourceName: "test-object",
			CallerIp:     "10.0.0.1",
		},
		JsonPayload: map[string]interface{}{
			"eventType": "OBJECT_CREATE",
			"size":      1024,
			"metadata": map[string]string{
				"contentType": "application/json",
			},
		},
	}

	// Test JSON marshaling (what would be sent to Kafka)
	data, err := json.Marshal(testLog)
	if err != nil {
		t.Fatalf("Failed to marshal audit log: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON data")
	}

	// Test unmarshaling
	var decoded models.AuditLog
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal audit log: %v", err)
	}

	// Verify key fields
	if decoded.InsertId != testLog.InsertId {
		t.Errorf("Expected InsertId %s, got %s", testLog.InsertId, decoded.InsertId)
	}

	if decoded.Severity != testLog.Severity {
		t.Errorf("Expected Severity %s, got %s", testLog.Severity, decoded.Severity)
	}

	if decoded.Resource.Type != testLog.Resource.Type {
		t.Errorf("Expected Resource.Type %s, got %s", testLog.Resource.Type, decoded.Resource.Type)
	}

	// Verify nested JsonPayload
	if decoded.JsonPayload["eventType"] != "OBJECT_CREATE" {
		t.Error("JsonPayload eventType not preserved")
	}

	if decoded.JsonPayload["size"].(float64) != 1024 {
		t.Error("JsonPayload size not preserved")
	}
}

func TestProcessor_GCSFileName(t *testing.T) {
	testLog := models.AuditLog{
		InsertId: "file-test-123",
	}

	// Test filename generation logic (what would be used for GCS)
	fileName := "audit-log-" + testLog.InsertId + "-" + time.Now().Format("2006-01-02-15-04-05") + ".json"

	if !contains(fileName, "audit-log-file-test-123") {
		t.Errorf("Filename should contain InsertId, got: %s", fileName)
	}

	if !contains(fileName, ".json") {
		t.Errorf("Filename should end with .json, got: %s", fileName)
	}

	// Test with bucket path
	bucketPath := "audit-logs"
	fullPath := bucketPath + "/" + fileName

	if !contains(fullPath, "audit-logs/audit-log-file-test-123") {
		t.Errorf("Full path should contain bucket path, got: %s", fullPath)
	}
}


