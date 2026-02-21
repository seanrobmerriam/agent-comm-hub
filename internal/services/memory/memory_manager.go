// Package memory provides memory management services.
package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"agent-comm-hub/internal/config"
	"agent-comm-hub/internal/models"
)

const (
	shortTermMemoryPrefix = "memory:short:"
	longTermMemoryPrefix  = "memory:long:"
)

// Errors for memory service.
var (
	ErrMemoryNotFound = errors.New("memory not found")
	ErrInvalidType    = errors.New("invalid memory type")
)

// MemoryManager handles agent memory operations.
type MemoryManager struct {
	httpClient *http.Client
	memoryURL  string
}

// NewMemoryManager creates a new memory manager.
func NewMemoryManager(cfg *config.MemoryConfig) *MemoryManager {
	return &MemoryManager{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		memoryURL: cfg.URL,
	}
}

// StoreShortTerm stores short-term memory.
func (m *MemoryManager) StoreShortTerm(ctx context.Context, agentID, key string, value interface{}, ttl time.Duration) error {
	// Store via HTTP to agent-memory-server
	reqBody := models.StoreMemoryRequest{
		MemoryType: models.MemoryTypeShortTerm,
		Key:        shortTermMemoryPrefix + agentID + ":" + key,
		Value:      value,
		TTL:        int(ttl.Seconds()),
	}

	return m.store(ctx, reqBody)
}

// GetShortTerm retrieves short-term memory.
func (m *MemoryManager) GetShortTerm(ctx context.Context, agentID, key string) (*models.Memory, error) {
	return m.get(ctx, shortTermMemoryPrefix+agentID+":"+key)
}

// DeleteShortTerm deletes short-term memory.
func (m *MemoryManager) DeleteShortTerm(ctx context.Context, agentID, key string) error {
	return m.delete(ctx, shortTermMemoryPrefix+agentID+":"+key)
}

// StoreLongTerm stores long-term memory.
func (m *MemoryManager) StoreLongTerm(ctx context.Context, agentID, key string, value interface{}) error {
	reqBody := models.StoreMemoryRequest{
		MemoryType: models.MemoryTypeLongTerm,
		Key:        longTermMemoryPrefix + agentID + ":" + key,
		Value:      value,
	}

	return m.store(ctx, reqBody)
}

// GetLongTerm retrieves long-term memory.
func (m *MemoryManager) GetLongTerm(ctx context.Context, agentID, key string) (*models.Memory, error) {
	return m.get(ctx, longTermMemoryPrefix+agentID+":"+key)
}

// DeleteLongTerm deletes long-term memory.
func (m *MemoryManager) DeleteLongTerm(ctx context.Context, agentID, key string) error {
	return m.delete(ctx, longTermMemoryPrefix+agentID+":"+key)
}

// SearchLongTerm searches long-term memory.
func (m *MemoryManager) SearchLongTerm(ctx context.Context, agentID, query string) ([]models.Memory, error) {
	// For now, return empty results - full-text search would require additional setup
	return []models.Memory{}, nil
}

func (m *MemoryManager) store(ctx context.Context, req models.StoreMemoryRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.memoryURL+"/memory", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to store memory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("memory server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (m *MemoryManager) get(ctx context.Context, key string) (*models.Memory, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", m.memoryURL+"/memory?key="+key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrMemoryNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("memory server returned status %d: %s", resp.StatusCode, string(body))
	}

	var memory models.Memory
	if err := json.NewDecoder(resp.Body).Decode(&memory); err != nil {
		return nil, fmt.Errorf("failed to decode memory: %w", err)
	}

	return &memory, nil
}

func (m *MemoryManager) delete(ctx context.Context, key string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", m.memoryURL+"/memory?key="+key, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("memory server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
