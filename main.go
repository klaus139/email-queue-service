package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"email-queue-service/config"
	"email-queue-service/handlers"
	"email-queue-service/service"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create email service
	emailService := service.NewEmailService(cfg.Workers, cfg.QueueSize)
	emailService.Start()

	// Create HTTP handler
	emailHandler := handlers.NewEmailHandler(emailService)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/send-email", emailHandler.SendEmailHandler)
	mux.HandleFunc("/dead-letter", emailHandler.DeadLetterHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown signal received")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Shutdown email service
	emailService.Shutdown()

	log.Println("Server exited")
}
