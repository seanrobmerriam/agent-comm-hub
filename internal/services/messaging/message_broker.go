// Package messaging provides message broker services.
package messaging

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
	directMessageChannelPrefix = "agent:message:"
	broadcastChannel           = "agent:broadcast"
	messageHistoryPrefix       = "agent:history:"
	messageHistoryTTL          = 24 * time.Hour // Messages kept for 24 hours
	messageHistoryMax          = 100            // Max messages per agent
)

// Errors for message broker.
var (
	ErrInvalidRecipient = errors.New("invalid recipient")
)

// MessageBroker handles message passing between agents.
type MessageBroker struct {
	redisPubSub *redis.Client
	redisStd    *redis.Client
}

// NewMessageBroker creates a new message broker.
func NewMessageBroker(redisPubSub, redisStd *redis.Client) *MessageBroker {
	return &MessageBroker{
		redisPubSub: redisPubSub,
		redisStd:    redisStd,
	}
}

// SendMessage sends a message to an agent.
func (b *MessageBroker) SendMessage(ctx context.Context, fromAgentID string, req *models.SendMessageRequest) (*models.Message, error) {
	// Create message
	msg := &models.Message{
		ID:            uuid.New().String(),
		FromAgent:     fromAgentID,
		ToAgent:       req.ToAgent,
		Type:          req.Type,
		Payload:       req.Payload,
		CorrelationID: req.CorrelationID,
		Timestamp:     time.Now(),
		TTL:           req.TTL,
	}

	// Serialize message
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Determine channel
	channel := directMessageChannelPrefix + req.ToAgent
	if req.ToAgent == "broadcast" {
		channel = broadcastChannel
	}

	// Publish message via Pub/Sub
	if err := b.redisPubSub.Publish(ctx, channel, data).Err(); err != nil {
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	// Store message history for sender
	if err := b.storeMessageHistory(ctx, fromAgentID, msg); err != nil {
		// Log error but don't fail the message send
		fmt.Printf("Warning: failed to store sender message history: %v\n", err)
	}

	// Store message history for receiver (if not broadcast)
	if req.ToAgent != "broadcast" {
		if err := b.storeMessageHistory(ctx, req.ToAgent, msg); err != nil {
			fmt.Printf("Warning: failed to store receiver message history: %v\n", err)
		}
	}

	return msg, nil
}

// GetMessageHistory retrieves message history for an agent.
func (b *MessageBroker) GetMessageHistory(ctx context.Context, agentID string, limit int) ([]models.Message, error) {
	if limit <= 0 || limit > messageHistoryMax {
		limit = messageHistoryMax
	}

	key := messageHistoryPrefix + agentID

	// Get messages from list (newest first)
	messages, err := b.redisStd.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	if len(messages) == 0 {
		return []models.Message{}, nil
	}

	// Parse messages (reverse to get chronological order)
	result := make([]models.Message, 0, len(messages))
	for i := len(messages) - 1; i >= 0; i-- {
		var msg models.Message
		if err := json.Unmarshal([]byte(messages[i]), &msg); err != nil {
			continue
		}
		result = append(result, msg)
	}

	return result, nil
}

// Subscribe subscribes to messages for an agent.
func (b *MessageBroker) Subscribe(ctx context.Context, agentID string) *redis.PubSub {
	channel := directMessageChannelPrefix + agentID
	return b.redisPubSub.Subscribe(ctx, channel)
}

// SubscribeToBroadcast subscribes to broadcast messages.
func (b *MessageBroker) SubscribeToBroadcast(ctx context.Context) *redis.PubSub {
	return b.redisPubSub.Subscribe(ctx, broadcastChannel)
}

func (b *MessageBroker) storeMessageHistory(ctx context.Context, agentID string, msg *models.Message) error {
	key := messageHistoryPrefix + agentID

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Add to list (LPUSH for newest first)
	if err := b.redisStd.LPush(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("failed to push message to history: %w", err)
	}

	// Trim list to max size
	if err := b.redisStd.LTrim(ctx, key, 0, messageHistoryMax-1).Err(); err != nil {
		return fmt.Errorf("failed to trim message history: %w", err)
	}

	// Set TTL on the key
	if err := b.redisStd.Expire(ctx, key, messageHistoryTTL).Err(); err != nil {
		return fmt.Errorf("failed to set message history TTL: %w", err)
	}

	return nil
}
