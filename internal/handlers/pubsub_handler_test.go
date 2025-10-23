package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"audit-logs-monitoring/internal/models"
	"audit-logs-monitoring/internal/services"
	"time"
)



func TestNewPubSubHandler(t *testing.T) {
	processor := services.NewProcessor(nil, nil)
	handler := NewPubSubHandler(processor)
	
	if handler == nil {
		t.Error("Expected handler, got nil")
	}
}

func TestHandlePubSub_ValidMessage(t *testing.T) {
	processor := services.NewProcessor(nil, nil)
	handler := NewPubSubHandler(processor)
	
	auditLog := models.AuditLog{
		Timestamp: time.Now(),
		Severity:  "INFO",
		LogName:   "test-log",
		Resource: models.Resource{
			Type: "compute",
		},
		ProtoPayload: &models.Payload{
			MethodName: "compute.instances.insert",
		},
	}
	
	logData, _ := json.Marshal(auditLog)
	
	pubsubMsg := PubSubMessage{
		Message: struct {
			Data       []byte            `json:"data"`
			Attributes map[string]string `json:"attributes"`
		}{
			Data: logData,
			Attributes: map[string]string{},
		},
	}
	
	msgData, _ := json.Marshal(pubsubMsg)
	
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(msgData))
	w := httptest.NewRecorder()
	
	handler.HandlePubSub(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Test passes if no error occurred
}

func TestHandlePubSub_InvalidMethod(t *testing.T) {
	processor := services.NewProcessor(nil, nil)
	handler := NewPubSubHandler(processor)
	
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	
	handler.HandlePubSub(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandlePubSub_InvalidJSON(t *testing.T) {
	processor := services.NewProcessor(nil, nil)
	handler := NewPubSubHandler(processor)
	
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	
	handler.HandlePubSub(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHealthCheck(t *testing.T) {
	processor := services.NewProcessor(nil, nil)
	handler := NewPubSubHandler(processor)
	
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	handler.HealthCheck(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "Healthy" {
		t.Errorf("Expected 'Healthy', got %s", w.Body.String())
	}
}
