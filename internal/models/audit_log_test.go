package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAuditLogSerialization(t *testing.T) {
	log := AuditLog{
		Timestamp: time.Now(),
		Severity:  "INFO",
		LogName:   "test-log",
		Resource: Resource{
			Type:   "compute",
			Labels: map[string]string{"zone": "us-central1"},
		},
		ProtoPayload: &Payload{
			Type:         "audit",
			MethodName:   "compute.instances.insert",
			ResourceName: "projects/test/zones/us-central1/instances/test-vm",
		},
	}

	data, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("Failed to marshal audit log: %v", err)
	}

	var unmarshaled AuditLog
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal audit log: %v", err)
	}

	if unmarshaled.Severity != log.Severity {
		t.Errorf("Expected severity %s, got %s", log.Severity, unmarshaled.Severity)
	}
	if unmarshaled.Resource.Type != log.Resource.Type {
		t.Errorf("Expected resource type %s, got %s", log.Resource.Type, unmarshaled.Resource.Type)
	}
}
