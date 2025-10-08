package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"audit-logs-monitoring/internal/models"
	"audit-logs-monitoring/pkg/logger"
)

type PubSubMessage struct {
	Message struct {
		Data       []byte            `json:"data"`
		Attributes map[string]string `json:"attributes"`
	} `json:"message"`
}

type LogProcessor interface {
	ProcessLog(log models.AuditLog)
}

type PubSubHandler struct {
	processor LogProcessor
}

func NewPubSubHandler(processor LogProcessor) *PubSubHandler {
	return &PubSubHandler{processor: processor}
}

func (h *PubSubHandler) HandlePubSub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg PubSubMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		logger.Infof("Error decoding Pub/Sub message: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var auditLog models.AuditLog
	if err := json.Unmarshal(msg.Message.Data, &auditLog); err != nil {
		logger.Infof("Error unmarshaling audit log: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	logger.Infof("Received Pub/Sub message: %+v", auditLog)
	h.processor.ProcessLog(auditLog)
	
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (h *PubSubHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Healthy")
}
