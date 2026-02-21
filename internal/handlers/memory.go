// Package handlers provides HTTP request handlers.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"agent-comm-hub/internal/models"
	"agent-comm-hub/internal/services/memory"
	"agent-comm-hub/internal/services/registry"
)

// MemoryHandler handles memory-related HTTP requests.
type MemoryHandler struct {
	memoryMgr *memory.MemoryManager
	registry  *registry.AgentRegistry
}

// NewMemoryHandler creates a new memory handler.
func NewMemoryHandler(memoryMgr *memory.MemoryManager, registry *registry.AgentRegistry) *MemoryHandler {
	return &MemoryHandler{
		memoryMgr: memoryMgr,
		registry:  registry,
	}
}

// Store handles POST /api/v1/agents/:id/memory - Store memory.
func (h *MemoryHandler) Store(w http.ResponseWriter, r *http.Request) {
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

	var req models.StoreMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	var storeErr error
	switch req.MemoryType {
	case models.MemoryTypeShortTerm:
		ttl := time.Duration(req.TTL) * time.Second
		if ttl == 0 {
			ttl = 1 * time.Hour // Default TTL: 1 hour
		}
		storeErr = h.memoryMgr.StoreShortTerm(r.Context(), agentID, req.Key, req.Value, ttl)
	case models.MemoryTypeLongTerm:
		storeErr = h.memoryMgr.StoreLongTerm(r.Context(), agentID, req.Key, req.Value)
	default:
		http.Error(w, "invalid memory_type (must be 'short_term' or 'long_term')", http.StatusBadRequest)
		return
	}

	if storeErr != nil {
		http.Error(w, storeErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.StoreMemoryResponse{
		Key:      req.Key,
		StoredAt: time.Now(),
	})
}

// Get handles GET /api/v1/agents/:id/memory - Retrieve memory.
func (h *MemoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	key := r.URL.Query().Get("key")
	memoryType := r.URL.Query().Get("type")

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

	if key != "" {
		// Get specific memory
		var mem *models.Memory
		var getErr error

		switch memoryType {
		case "short_term":
			mem, getErr = h.memoryMgr.GetShortTerm(r.Context(), agentID, key)
		case "long_term", "":
			mem, getErr = h.memoryMgr.GetLongTerm(r.Context(), agentID, key)
		default:
			http.Error(w, "invalid memory type", http.StatusBadRequest)
			return
		}

		if errors.Is(getErr, memory.ErrMemoryNotFound) {
			http.Error(w, "memory not found", http.StatusNotFound)
			return
		}
		if getErr != nil {
			http.Error(w, getErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mem)
		return
	}

	// For now, list is not implemented - would require additional API support
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.MemoryListResponse{
		Memories: []models.Memory{},
		Count:    0,
	})
}

// Delete handles DELETE /api/v1/agents/:id/memory - Delete memory.
func (h *MemoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	key := r.URL.Query().Get("key")
	memoryType := r.URL.Query().Get("type")

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

	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	var deleteErr error
	switch memoryType {
	case "short_term":
		deleteErr = h.memoryMgr.DeleteShortTerm(r.Context(), agentID, key)
	case "long_term", "":
		deleteErr = h.memoryMgr.DeleteLongTerm(r.Context(), agentID, key)
	default:
		http.Error(w, "invalid memory type", http.StatusBadRequest)
		return
	}

	if deleteErr != nil {
		http.Error(w, deleteErr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
