// Package models provides data models for the application.
package models

import (
	"time"
)

// MemoryType represents the type of memory.
type MemoryType string

const (
	MemoryTypeShortTerm MemoryType = "short_term"
	MemoryTypeLongTerm  MemoryType = "long_term"
)

// Memory represents agent memory.
type Memory struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	MemoryType MemoryType  `json:"memory_type"`
	StoredAt   time.Time   `json:"stored_at"`
	TTL        int         `json:"ttl,omitempty"` // TTL in seconds for short-term memory
}

// StoreMemoryRequest represents a request to store memory.
type StoreMemoryRequest struct {
	MemoryType MemoryType  `json:"memory_type" validate:"required"`
	Key        string      `json:"key" validate:"required"`
	Value      interface{} `json:"value" validate:"required"`
	TTL        int         `json:"ttl"` // TTL in seconds for short-term memory
}

// StoreMemoryResponse represents the response after storing memory.
type StoreMemoryResponse struct {
	Key      string    `json:"key"`
	StoredAt time.Time `json:"stored_at"`
}

// MemoryListResponse represents a list of memories.
type MemoryListResponse struct {
	Memories []Memory `json:"memories"`
	Count    int      `json:"count"`
}
