// Package handlers provides HTTP request handlers.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"agent-comm-hub/internal/models"
	"agent-comm-hub/internal/services/registry"
)

// AgentHandler handles agent-related HTTP requests.
type AgentHandler struct {
	registry *registry.AgentRegistry
}

// NewAgentHandler creates a new agent handler.
func NewAgentHandler(registry *registry.AgentRegistry) *AgentHandler {
	return &AgentHandler{
		registry: registry,
	}
}

// Register handles POST /api/v1/agents - Register a new agent.
func (h *AgentHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.Type == "" {
		http.Error(w, "name and type are required", http.StatusBadRequest)
		return
	}

	agent, err := h.registry.Register(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.RegisterAgentResponse{
		ID:     agent.ID,
		Status: agent.Status,
	})
}

// List handles GET /api/v1/agents - List all agents.
func (h *AgentHandler) List(w http.ResponseWriter, r *http.Request) {
	agents, err := h.registry.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AgentListResponse{
		Agents: agents,
		Count:  len(agents),
	})
}

// Get handles GET /api/v1/agents/:id - Get agent details.
func (h *AgentHandler) Get(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	agent, err := h.registry.Get(r.Context(), agentID)
	if err != nil {
		if errors.Is(err, registry.ErrAgentNotFound) {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// Update handles PUT /api/v1/agents/:id - Update agent.
func (h *AgentHandler) Update(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	var req models.UpdateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	agent, err := h.registry.Update(r.Context(), agentID, &req)
	if err != nil {
		if errors.Is(err, registry.ErrAgentNotFound) {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// Delete handles DELETE /api/v1/agents/:id - Unregister agent.
func (h *AgentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	err := h.registry.Unregister(r.Context(), agentID)
	if err != nil {
		if errors.Is(err, registry.ErrAgentNotFound) {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Heartbeat handles POST /api/v1/agents/:id/heartbeat - Agent heartbeat.
func (h *AgentHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	err := h.registry.Heartbeat(r.Context(), agentID)
	if err != nil {
		if errors.Is(err, registry.ErrAgentNotFound) {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
