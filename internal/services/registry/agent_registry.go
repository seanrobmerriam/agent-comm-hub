// Package registry provides agent registry services.
package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"agent-comm-hub/internal/models"
)

const (
	agentKeyPrefix    = "agent:"
	agentIndexKey     = "agents:index"
	agentHeartbeatTTL = 5 * time.Minute
)

// Errors for agent registry.
var (
	ErrAgentNotFound = errors.New("agent not found")
	ErrAgentExists   = errors.New("agent already exists")
)

// AgentRegistry manages agent registration and discovery.
type AgentRegistry struct {
	redis *redis.Client
}

// NewAgentRegistry creates a new agent registry.
func NewAgentRegistry(redisClient *redis.Client) *AgentRegistry {
	return &AgentRegistry{
		redis: redisClient,
	}
}

// Register registers a new agent.
func (r *AgentRegistry) Register(ctx context.Context, req *models.RegisterAgentRequest) (*models.Agent, error) {
	// Generate unique ID
	agentID := uuid.New().String()

	now := time.Now()
	agent := &models.Agent{
		ID:           agentID,
		Name:         req.Name,
		Type:         req.Type,
		Capabilities: req.Capabilities,
		Endpoint:     req.Endpoint,
		Status:       models.StatusOnline,
		Metadata:     req.Metadata,
		CreatedAt:    now,
		LastSeen:     now,
	}

	// Store agent data
	agentKey := agentKeyPrefix + agentID
	data, err := json.Marshal(agent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent: %w", err)
	}

	if err := r.redis.Set(ctx, agentKey, data, 0).Err(); err != nil {
		return nil, fmt.Errorf("failed to store agent: %w", err)
	}

	// Add to index
	if err := r.redis.SAdd(ctx, agentIndexKey, agentID).Err(); err != nil {
		return nil, fmt.Errorf("failed to add agent to index: %w", err)
	}

	// Set heartbeat
	if err := r.updateHeartbeat(ctx, agentID); err != nil {
		return nil, err
	}

	return agent, nil
}

// Get retrieves an agent by ID.
func (r *AgentRegistry) Get(ctx context.Context, agentID string) (*models.Agent, error) {
	agentKey := agentKeyPrefix + agentID
	data, err := r.redis.Get(ctx, agentKey).Bytes()
	if err == redis.Nil {
		return nil, ErrAgentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	var agent models.Agent
	if err := json.Unmarshal(data, &agent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent: %w", err)
	}

	return &agent, nil
}

// List retrieves all registered agents.
func (r *AgentRegistry) List(ctx context.Context) ([]models.Agent, error) {
	// Get all agent IDs from index
	agentIDs, err := r.redis.SMembers(ctx, agentIndexKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent index: %w", err)
	}

	if len(agentIDs) == 0 {
		return []models.Agent{}, nil
	}

	// Get all agents
	agents := make([]models.Agent, 0, len(agentIDs))
	for _, agentID := range agentIDs {
		agent, err := r.Get(ctx, agentID)
		if err != nil {
			// Skip agents that can't be retrieved
			continue
		}
		agents = append(agents, *agent)
	}

	return agents, nil
}

// Update updates an existing agent.
func (r *AgentRegistry) Update(ctx context.Context, agentID string, req *models.UpdateAgentRequest) (*models.Agent, error) {
	// Get existing agent
	agent, err := r.Get(ctx, agentID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != "" {
		agent.Name = req.Name
	}
	if req.Type != "" {
		agent.Type = req.Type
	}
	if len(req.Capabilities) > 0 {
		agent.Capabilities = req.Capabilities
	}
	if req.Endpoint != "" {
		agent.Endpoint = req.Endpoint
	}
	if req.Status != "" {
		agent.Status = req.Status
	}
	if req.Metadata != nil {
		agent.Metadata = req.Metadata
	}

	// Save updated agent
	agentKey := agentKeyPrefix + agentID
	data, err := json.Marshal(agent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent: %w", err)
	}

	if err := r.redis.Set(ctx, agentKey, data, 0).Err(); err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	return agent, nil
}

// Unregister removes an agent from the registry.
func (r *AgentRegistry) Unregister(ctx context.Context, agentID string) error {
	// Check if agent exists
	_, err := r.Get(ctx, agentID)
	if err != nil {
		return err
	}

	// Remove from index
	if err := r.redis.SRem(ctx, agentIndexKey, agentID).Err(); err != nil {
		return fmt.Errorf("failed to remove agent from index: %w", err)
	}

	// Delete agent data
	agentKey := agentKeyPrefix + agentID
	if err := r.redis.Del(ctx, agentKey).Err(); err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	return nil
}

// Heartbeat updates the agent's last seen timestamp.
func (r *AgentRegistry) Heartbeat(ctx context.Context, agentID string) error {
	// Check if agent exists
	_, err := r.Get(ctx, agentID)
	if err != nil {
		return err
	}

	return r.updateHeartbeat(ctx, agentID)
}

func (r *AgentRegistry) updateHeartbeat(ctx context.Context, agentID string) error {
	heartbeatKey := "agent:heartbeat:" + agentID
	if err := r.redis.Set(ctx, heartbeatKey, time.Now().Unix(), agentHeartbeatTTL).Err(); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	// Update last seen in agent data
	agentKey := agentKeyPrefix + agentID
	data, err := r.redis.Get(ctx, agentKey).Bytes()
	if err != nil {
		return fmt.Errorf("failed to get agent for heartbeat: %w", err)
	}

	var agent models.Agent
	if err := json.Unmarshal(data, &agent); err != nil {
		return fmt.Errorf("failed to unmarshal agent: %w", err)
	}

	agent.LastSeen = time.Now()

	updatedData, err := json.Marshal(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent: %w", err)
	}

	if err := r.redis.Set(ctx, agentKey, updatedData, 0).Err(); err != nil {
		return fmt.Errorf("failed to save agent after heartbeat: %w", err)
	}

	return nil
}
