package services

import (
	"strings"
	"audit-logs-monitoring/internal/models"
)

type IntrusionDetector struct{}

func NewIntrusionDetector() *IntrusionDetector {
	return &IntrusionDetector{}
}

func (id *IntrusionDetector) DetectIntrusion(log *models.AuditLog) *models.IntrusionAlert {
	if log.ProtoPayload == nil {
		return nil
	}

	// Network intrusion patterns
	suspiciousPatterns := []string{
		"compute.firewalls.insert",
		"compute.firewalls.patch",
		"compute.networks.insert",
		"compute.subnetworks.insert",
		"iam.serviceAccounts.create",
		"storage.buckets.setIamPolicy",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(log.ProtoPayload.MethodName, pattern) {
			return &models.IntrusionAlert{
				Timestamp:   log.Timestamp,
				Severity:    "HIGH",
				Description: "Suspicious network configuration change detected",
				SourceIP:    extractSourceIP(log),
				Resource:    log.ProtoPayload.ResourceName,
				Action:      log.ProtoPayload.MethodName,
			}
		}
	}

	// Check for unusual access patterns
	if log.Severity == "ERROR" && strings.Contains(log.ProtoPayload.MethodName, "compute") {
		return &models.IntrusionAlert{
			Timestamp:   log.Timestamp,
			Severity:    "MEDIUM",
			Description: "Failed compute operation - potential intrusion attempt",
			SourceIP:    extractSourceIP(log),
			Resource:    log.ProtoPayload.ResourceName,
			Action:      log.ProtoPayload.MethodName,
		}
	}

	return nil
}

func extractSourceIP(log *models.AuditLog) string {
	if log.ProtoPayload == nil {
		return "unknown"
	}

	// Try callerIp first
	if log.ProtoPayload.CallerIp != "" {
		return log.ProtoPayload.CallerIp
	}

	// Try requestMetadata
	if log.ProtoPayload.RequestMetadata != nil && log.ProtoPayload.RequestMetadata.CallerIp != "" {
		return log.ProtoPayload.RequestMetadata.CallerIp
	}

	// Try request field
	if request, ok := log.ProtoPayload.Request["sourceIP"]; ok {
		if ip, ok := request.(string); ok {
			return ip
		}
	}

	// Try httpRequest
	if log.HttpRequest != nil && log.HttpRequest.RemoteIp != "" {
		return log.HttpRequest.RemoteIp
	}

	return "unknown"
}