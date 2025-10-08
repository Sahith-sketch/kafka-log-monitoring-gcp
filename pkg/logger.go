package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	level := logrus.InfoLevel
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		if parsedLevel, err := logrus.ParseLevel(strings.ToLower(envLevel)); err == nil {
			level = parsedLevel
		}
	}
	log.SetLevel(level)
}

func Info(msg string) {
	log.Info(msg)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Error(msg string) {
	log.Error(msg)
}

func WithError(err error) *logrus.Entry {
	return log.WithError(err)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return log.WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}
