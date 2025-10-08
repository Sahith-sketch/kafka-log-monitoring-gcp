package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                   string
	BatchSize              int
	BatchTime              time.Duration
	WorkerCount            int
	KafkaBrokers           []string
	KafkaTopic             string
	GCSBucket              string
	BucketPath             string
	TLSCACertSecret        string
	TLSClientCertSecret    string
	TLSClientKeySecret     string
	ProjectID              string
	SMTPHost               string
	SMTPPort               int
	SMTPUser               string
	SMTPPass               string
	AlertEmail             string
	IsKafkaEnabled         bool
	IsGCSEnabled           bool
	IsNotificationsEnabled bool
}

func Load() *Config {
	batchSize, _ := strconv.Atoi(getEnv("BATCH_SIZE", "100"))
	batchTimeMs, _ := strconv.Atoi(getEnv("BATCH_TIME", "5000"))
	workerCount, _ := strconv.Atoi(getEnv("WORKER_COUNT", "10"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		Port:                   getEnv("PORT", "8080"),
		BatchSize:              batchSize,
		BatchTime:              time.Duration(batchTimeMs) * time.Millisecond,
		WorkerCount:            workerCount,
		KafkaBrokers:           []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopic:             getEnv("KAFKA_TOPIC", "audit-logs"),
		GCSBucket:              getEnv("BUCKET_NAME", "audit-logs-storage"),
		BucketPath:             getEnv("BUCKET_PATH", ""),
		TLSCACertSecret:        getEnv("TLS_CA_CERT_SECRET", ""),
		TLSClientCertSecret:    getEnv("TLS_CLIENT_CERT_SECRET", ""),
		TLSClientKeySecret:     getEnv("TLS_CLIENT_KEY_SECRET", ""),
		ProjectID:              getEnv("PROJECT_ID", ""),
		IsKafkaEnabled:         getBoolEnv("IS_KAFKA_ENABLED", false),
		IsGCSEnabled:           getBoolEnv("IS_GCS_ENABLED", false),
		IsNotificationsEnabled: getBoolEnv("IS_NOTIFICATIONS_ENABLED", false),
		SMTPHost:               getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:               smtpPort,
		SMTPUser:               getEnv("SMTP_USER", ""),
		SMTPPass:               getEnv("SMTP_PASS", ""),
		AlertEmail:             getEnv("ALERT_EMAIL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
