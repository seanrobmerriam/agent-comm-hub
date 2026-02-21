package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"agent-comm-hub/internal/config"
	"agent-comm-hub/internal/handlers"
	"agent-comm-hub/internal/services/memory"
	"agent-comm-hub/internal/services/messaging"
	"agent-comm-hub/internal/services/redis"
	"agent-comm-hub/internal/services/registry"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Redis manager
	redisManager, err := redis.NewManager(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redisManager.Close()
	log.Println("Redis connections established")

	// Initialize services
	agentRegistry := registry.NewAgentRegistry(redisManager.Standard())
	messageBroker := messaging.NewMessageBroker(redisManager.PubSub(), redisManager.Standard())
	memoryManager := memory.NewMemoryManager(&cfg.Memory)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(redisManager)
	agentHandler := handlers.NewAgentHandler(agentRegistry)
	messageHandler := handlers.NewMessageHandler(messageBroker, agentRegistry)
	memoryHandler := handlers.NewMemoryHandler(memoryManager, agentRegistry)

	// Setup router
	router := setupRouter(healthHandler, agentHandler, messageHandler, memoryHandler)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

func setupRouter(healthHandler *handlers.HealthHandler, agentHandler *handlers.AgentHandler, messageHandler *handlers.MessageHandler, memoryHandler *handlers.MemoryHandler) *chi.Mux {
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// Health endpoints
	router.Get("/health", healthHandler.Handle)
	router.Get("/ready", healthHandler.Ready)

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		// Agent routes
		r.Route("/agents", func(r chi.Router) {
			r.Post("/", agentHandler.Register)
			r.Get("/", agentHandler.List)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", agentHandler.Get)
				r.Put("/", agentHandler.Update)
				r.Delete("/", agentHandler.Delete)
				r.Post("/heartbeat", agentHandler.Heartbeat)
				// Message routes
				r.Route("/messages", func(r chi.Router) {
					r.Post("/", messageHandler.Send)
					r.Get("/", messageHandler.List)
				})
				// Memory routes
				r.Route("/memory", func(r chi.Router) {
					r.Post("/", memoryHandler.Store)
					r.Get("/", memoryHandler.Get)
					r.Delete("/", memoryHandler.Delete)
				})
			})
		})
	})

	return router
}
