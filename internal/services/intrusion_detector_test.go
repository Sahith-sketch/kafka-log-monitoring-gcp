package services

import (
	"testing"
	"audit-logs-monitoring/internal/models"
	"time"
)

func TestDetectIntrusion_SuspiciousPatterns(t *testing.T) {
	detector := NewIntrusionDetector()
	
	testCases := []struct {
		methodName string
		shouldAlert bool
	}{
		{"compute.firewalls.insert", true},
		{"compute.firewalls.patch", true},
		{"compute.networks.insert", true},
		{"iam.serviceAccounts.create", true},
		{"compute.instances.insert", false},
		{"storage.objects.get", false},
	}
	
	for _, tc := range testCases {
		log := &models.AuditLog{
			Timestamp: time.Now(),
			Severity:  "INFO",
			ProtoPayload: &models.Payload{
				MethodName: tc.methodName,
				ResourceName: "test-resource",
			},
		}
		
		alert := detector.DetectIntrusion(log)
		if tc.shouldAlert && alert == nil {
			t.Errorf("Expected alert for method %s, got none", tc.methodName)
		}
		if !tc.shouldAlert && alert != nil {
			t.Errorf("Expected no alert for method %s, got alert", tc.methodName)
		}
		if alert != nil && alert.Severity != "HIGH" {
			t.Errorf("Expected HIGH severity, got %s", alert.Severity)
		}
	}
}

func TestDetectIntrusion_ErrorPatterns(t *testing.T) {
	detector := NewIntrusionDetector()
	
	log := &models.AuditLog{
		Timestamp: time.Now(),
		Severity:  "ERROR",
		ProtoPayload: &models.Payload{
			MethodName: "compute.instances.delete",
			ResourceName: "test-resource",
		},
	}
	
	alert := detector.DetectIntrusion(log)
	if alert == nil {
		t.Error("Expected alert for ERROR severity compute operation")
	}
	if alert != nil && alert.Severity != "MEDIUM" {
		t.Errorf("Expected MEDIUM severity, got %s", alert.Severity)
	}
}

func TestDetectIntrusion_NoAlert(t *testing.T) {
	detector := NewIntrusionDetector()
	
	log := &models.AuditLog{
		Timestamp: time.Now(),
		Severity:  "INFO",
		ProtoPayload: &models.Payload{
			MethodName: "storage.objects.get",
			ResourceName: "test-resource",
		},
	}
	
	alert := detector.DetectIntrusion(log)
	if alert != nil {
		t.Error("Expected no alert for normal operation")
	}
}

func TestExtractSourceIP(t *testing.T) {
	log := &models.AuditLog{
		ProtoPayload: &models.Payload{
			Request: map[string]interface{}{
				"sourceIP": "192.168.1.1",
			},
		},
	}
	
	ip := extractSourceIP(log)
	if ip != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", ip)
	}
	
	logNoIP := &models.AuditLog{
		ProtoPayload: &models.Payload{
			Request: map[string]interface{}{},
		},
	}
	
	ip = extractSourceIP(logNoIP)
	if ip != "unknown" {
		t.Errorf("Expected unknown IP, got %s", ip)
	}
}