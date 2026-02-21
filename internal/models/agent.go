// Package models provides data models for the application.
package models

import (
	"time"
)

// AgentStatus represents the status of an agent.
type AgentStatus string

const (
	StatusOnline  AgentStatus = "online"
	StatusOffline AgentStatus = "offline"
	StatusBusy    AgentStatus = "busy"
)

// Agent represents a registered agent.
type Agent struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Capabilities []string          `json:"capabilities"`
	Endpoint     string            `json:"endpoint"`
	Status       AgentStatus       `json:"status"`
	Metadata     map[string]string `json:"metadata"`
	CreatedAt    time.Time         `json:"created_at"`
	LastSeen     time.Time         `json:"last_seen"`
}

// RegisterAgentRequest represents a request to register an agent.
type RegisterAgentRequest struct {
	Name         string            `json:"name" validate:"required"`
	Type         string            `json:"type" validate:"required"`
	Capabilities []string          `json:"capabilities"`
	Endpoint     string            `json:"endpoint"`
	Metadata     map[string]string `json:"metadata"`
}

// RegisterAgentResponse represents the response after agent registration.
type RegisterAgentResponse struct {
	ID     string      `json:"id"`
	Status AgentStatus `json:"status"`
}

// UpdateAgentRequest represents a request to update an agent.
type UpdateAgentRequest struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Capabilities []string          `json:"capabilities"`
	Endpoint     string            `json:"endpoint"`
	Status       AgentStatus       `json:"status"`
	Metadata     map[string]string `json:"metadata"`
}

// AgentListResponse represents a list of agents response.
type AgentListResponse struct {
	Agents []Agent `json:"agents"`
	Count  int     `json:"count"`
}
