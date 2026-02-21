# Agent Communication Hub

A Go-based system that enables independent agentic AI instances to discover, communicate, and collaborate with each other.

## Overview

The Agent Communication Hub provides:
- **Agent Registration & Discovery** - Agents can register themselves and discover other available agents
- **Real-time Message Passing** - Redis pub/sub enables instant agent-to-agent communication
- **Persistent Memory** - Agent-memory-server provides long-term and short-term memory storage
- **Standard Data Storage** - Redis handles caching, session management, and metadata

## Architecture

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
│   Agents    │────▶│  API Gateway    │────▶│    Redis    │
│             │     │    (Port 8080)  │     │  (Standard) │
└─────────────┘     └─────────────────┘     └─────────────┘
                           │
                           ▼
                   ┌─────────────────┐
                   │  Redis Pub/Sub  │
                   │    (Port 6380)  │
                   └─────────────────┘
                           │
                           ▼
                   ┌─────────────────────┐
                   │ Agent Memory Server │
                   │     (Port 8081)     │
                   └─────────────────────┘
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### Running with Docker Compose

```bash
cd agent-comm-hub
docker-compose up -d
```

This will start:
- `agent-comm-hub` on port 8080
- `redis-standard` on port 6379
- `redis-pubsub` on port 6380
- `agent-memory-server` on port 8081

### Running Locally

```bash
# Copy environment configuration
cp .env.example .env

# Build the application
go build -o agent-comm-hub ./cmd/server

# Run the server
./agent-comm-hub
```

## API Endpoints

### Health Check
- `GET /health` - Service health check
- `GET /ready` - Readiness check

### Agent Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/v1/agents | Register a new agent |
| GET | /api/v1/agents | List all agents |
| GET | /api/v1/agents/:id | Get agent details |
| PUT | /api/v1/agents/:id | Update agent |
| DELETE | /api/v1/agents/:id | Unregister agent |
| POST | /api/v1/agents/:id/heartbeat | Agent heartbeat |

### Messaging
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/v1/agents/:id/messages | Send message |
| GET | /api/v1/agents/:id/messages | Get message history |

### Memory
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/v1/agents/:id/memory | Store memory |
| GET | /api/v1/agents/:id/memory | Retrieve memory |
| DELETE | /api/v1/agents/:id/memory | Delete memory |

## Example Usage

### Register an Agent

```bash
curl -X POST http://localhost:8080/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Research Agent",
    "type": "research",
    "capabilities": ["web-search", "data-analysis"],
    "endpoint": "http://agent:8080",
    "metadata": {"version": "1.0.0"}
  }'
```

### Send a Message

```bash
curl -X POST http://localhost:8080/api/v1/agents/{agent-id}/messages \
  -H "Content-Type: application/json" \
  -d '{
    "to_agent": "target-agent-id",
    "type": "request",
    "payload": {"action": "analyze", "data": "..."},
    "ttl": 300
  }'
```

### Store Memory

```bash
curl -X POST http://localhost:8080/api/v1/agents/{agent-id}/memory \
  -H "Content-Type: application/json" \
  -d '{
    "memory_type": "long_term",
    "key": "knowledge:project-x",
    "value": {"description": "Project X details", "data": "..."}
  }'
```

## Configuration

Environment variables can be configured in `.env`:

| Variable | Default | Description |
|----------|---------|-------------|
| SERVER_HOST | 0.0.0.0 | Server host |
| SERVER_PORT | 8080 | Server port |
| REDIS_STANDARD_URL | redis://localhost:6379 | Standard Redis URL |
| REDIS_PUBSUB_URL | redis://localhost:6380 | Pub/Sub Redis URL |
| AGENT_MEMORY_URL | http://localhost:8081 | Agent Memory Server URL |
| LOG_LEVEL | info | Logging level |

## Project Structure

```
agent-comm-hub/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── handlers/             # HTTP handlers
│   ├── middleware/          # HTTP middleware
│   ├── models/              # Data models
│   └── services/            # Business logic services
│       ├── memory/           # Memory management
│       ├── messaging/       # Message broker
│       ├── redis/           # Redis connections
│       └── registry/        # Agent registry
├── .env.example             # Environment configuration template
├── docker-compose.yml       # Docker Compose configuration
├── Dockerfile              # Application Dockerfile
├── go.mod                  # Go module definition
└── README.md               # This file
```

## License

MIT
