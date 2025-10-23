package handlers

import (
	"audit-logs-monitoring/internal/models"
	"audit-logs-monitoring/internal/services"
	"audit-logs-monitoring/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"
)

type PubSubMessage struct {
	Message struct {
		Data       []byte            `json:"data"`
		Attributes map[string]string `json:"attributes"`
	} `json:"message"`
}

type PubSubHandler struct {
	processor *services.Processor
}

func NewPubSubHandler(processor *services.Processor) *PubSubHandler {
	return &PubSubHandler{processor: processor}
}

func (h *PubSubHandler) HandlePubSub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg PubSubMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		logger.WithError(err).Error("Error decoding Pub/Sub message")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var auditLog models.AuditLog
	if err := json.Unmarshal(msg.Message.Data, &auditLog); err != nil {
		logger.WithError(err).Error("Error unmarshaling audit log")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := h.processor.ProcessLog(r.Context(), auditLog); err != nil {
		logger.WithError(err).Error("Error processing audit log")
		http.Error(w, "Processing failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (h *PubSubHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Healthy")
}