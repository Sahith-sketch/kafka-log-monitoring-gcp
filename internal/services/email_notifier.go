package services

import (
	"fmt"
	"audit-logs-monitoring/internal/config"
	"audit-logs-monitoring/internal/models"

	"gopkg.in/gomail.v2"
)

type EmailNotifier struct {
	config *config.Config
}

func NewEmailNotifier(cfg *config.Config) *EmailNotifier {
	return &EmailNotifier{config: cfg}
}

func (en *EmailNotifier) SendAlert(alert *models.IntrusionAlert) error {
	if en.config.AlertEmail == "" || en.config.SMTPUser == "" {
		return fmt.Errorf("email configuration missing")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", en.config.SMTPUser)
	m.SetHeader("To", en.config.AlertEmail)
	m.SetHeader("Subject", fmt.Sprintf("Security Alert: %s", alert.Severity))
	
	body := fmt.Sprintf(`
Security Intrusion Detected

Timestamp: %s
Severity: %s
Description: %s
Source IP: %s
Resource: %s
Action: %s
`, 
		alert.Timestamp.Format("2006-01-02 15:04:05"),
		alert.Severity,
		alert.Description,
		alert.SourceIP,
		alert.Resource,
		alert.Action,
	)
	
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(en.config.SMTPHost, en.config.SMTPPort, en.config.SMTPUser, en.config.SMTPPass)
	return d.DialAndSend(m)
}