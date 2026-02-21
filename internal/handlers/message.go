// Package handlers provides HTTP request handlers.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"agent-comm-hub/internal/models"
	"agent-comm-hub/internal/services/messaging"
	"agent-comm-hub/internal/services/registry"
)

// MessageHandler handles message-related HTTP requests.
type MessageHandler struct {
	broker   *messaging.MessageBroker
	registry *registry.AgentRegistry
}

// NewMessageHandler creates a new message handler.
func NewMessageHandler(broker *messaging.MessageBroker, registry *registry.AgentRegistry) *MessageHandler {
	return &MessageHandler{
		broker:   broker,
		registry: registry,
	}
}

// Send handles POST /api/v1/agents/:id/messages - Send a message.
func (h *MessageHandler) Send(w http.ResponseWriter, r *http.Request) {
	fromAgentID := chi.URLParam(r, "id")

	// Verify sender exists
	_, err := h.registry.Get(r.Context(), fromAgentID)
	if errors.Is(err, registry.ErrAgentNotFound) {
		http.Error(w, "sender agent not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ToAgent == "" {
		http.Error(w, "to_agent is required", http.StatusBadRequest)
		return
	}

	// Set default message type
	if req.Type == "" {
		req.Type = models.MessageTypeMessage
	}

	msg, err := h.broker.SendMessage(r.Context(), fromAgentID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(models.SendMessageResponse{
		MessageID: msg.ID,
		Timestamp: msg.Timestamp,
		Channel:   "agent:message:" + req.ToAgent,
	})
}

// List handles GET /api/v1/agents/:id/messages - Get message history.
func (h *MessageHandler) List(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	// Verify agent exists
	_, err := h.registry.Get(r.Context(), agentID)
	if errors.Is(err, registry.ErrAgentNotFound) {
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get limit from query param
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	messages, err := h.broker.GetMessageHistory(r.Context(), agentID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.MessageListResponse{
		Messages: messages,
		Count:    len(messages),
	})
}
