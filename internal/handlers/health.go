// Package handlers provides HTTP request handlers.
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"agent-comm-hub/internal/services/redis"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	redisManager *redis.Manager
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(redisManager *redis.Manager) *HealthHandler {
	return &HealthHandler{
		redisManager: redisManager,
	}
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// Handle handles health check requests.
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  make(map[string]string),
	}

	// Check Redis connections
	if h.redisManager != nil {
		if err := h.redisManager.Ping(ctx); err != nil {
			response.Status = "degraded"
			response.Services["redis"] = "unhealthy: " + err.Error()
		} else {
			response.Services["redis"] = "healthy"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Ready checks if the service is ready to accept traffic.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check Redis
	if h.redisManager != nil {
		if err := h.redisManager.Ping(ctx); err != nil {
			http.Error(w, "service not ready", http.StatusServiceUnavailable)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
