package services

import (
	"testing"
	"audit-logs-monitoring/internal/config"
	"audit-logs-monitoring/internal/models"
	"time"
)

func TestNewEmailNotifier(t *testing.T) {
	cfg := &config.Config{
		SMTPHost: "smtp.test.com",
		SMTPPort: 587,
		SMTPUser: "test@test.com",
		SMTPPass: "password",
		AlertEmail: "alert@test.com",
	}
	
	notifier := NewEmailNotifier(cfg)
	if notifier == nil {
		t.Error("Expected email notifier, got nil")
	}
	if notifier.config != cfg {
		t.Error("Config not properly set")
	}
}

func TestSendAlert_MissingConfig(t *testing.T) {
	cfg := &config.Config{}
	notifier := NewEmailNotifier(cfg)
	
	alert := &models.IntrusionAlert{
		Timestamp: time.Now(),
		Severity: "HIGH",
		Description: "Test alert",
	}
	
	err := notifier.SendAlert(alert)
	if err == nil {
		t.Error("Expected error for missing email config")
	}
}

func TestSendAlert_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		SMTPHost: "smtp.test.com",
		SMTPPort: 587,
		SMTPUser: "test@test.com",
		SMTPPass: "password",
		AlertEmail: "alert@test.com",
	}
	
	notifier := NewEmailNotifier(cfg)
	alert := &models.IntrusionAlert{
		Timestamp: time.Now(),
		Severity: "HIGH",
		Description: "Test alert",
		SourceIP: "192.168.1.1",
		Resource: "test-resource",
		Action: "test-action",
	}
	
	// This will fail to send but should not panic
	err := notifier.SendAlert(alert)
	// We expect an error since we're not connecting to a real SMTP server
	if err == nil {
		t.Log("Warning: Expected SMTP connection error, but got none")
	}
}