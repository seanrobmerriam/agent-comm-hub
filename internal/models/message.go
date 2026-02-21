// Package models provides data models for the application.
package models

import (
	"time"
)

// MessageType represents the type of message.
type MessageType string

const (
	MessageTypeRequest  MessageType = "request"
	MessageTypeResponse MessageType = "response"
	MessageTypeEvent    MessageType = "event"
	MessageTypeMessage  MessageType = "message"
)

// Message represents a message between agents.
type Message struct {
	ID            string      `json:"id"`
	FromAgent     string      `json:"from_agent"`
	ToAgent       string      `json:"to_agent"`
	Type          MessageType `json:"type"`
	Payload       interface{} `json:"payload"`
	CorrelationID string      `json:"correlation_id,omitempty"`
	Timestamp     time.Time   `json:"timestamp"`
	TTL           int         `json:"ttl,omitempty"` // TTL in seconds, 0 = no expiration
}

// SendMessageRequest represents a request to send a message.
type SendMessageRequest struct {
	ToAgent       string      `json:"to_agent" validate:"required"`
	Type          MessageType `json:"type"`
	Payload       interface{} `json:"payload"`
	CorrelationID string      `json:"correlation_id"`
	TTL           int         `json:"ttl"`
}

// SendMessageResponse represents the response after sending a message.
type SendMessageResponse struct {
	MessageID string    `json:"message_id"`
	Timestamp time.Time `json:"timestamp"`
	Channel   string    `json:"channel"`
}

// MessageListResponse represents a list of messages.
type MessageListResponse struct {
	Messages []Message `json:"messages"`
	Count    int       `json:"count"`
}
