package handlers

import (
	"encoding/json"
	"net/http"

	"email-queue-service/models"
	"email-queue-service/service"
	"email-queue-service/utils"
)

// EmailHandler handles email-related HTTP requests
type EmailHandler struct {
	emailService *service.EmailService
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(emailService *service.EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

// SendEmailHandler handles POST /send-email requests
func (h *EmailHandler) SendEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.To == "" || req.Subject == "" || req.Body == "" {
		http.Error(w, "All fields (to, subject, body) are required", http.StatusUnprocessableEntity)
		return
	}

	// Validate email format
	if !utils.ValidateEmail(req.To) {
		http.Error(w, "Invalid email format", http.StatusUnprocessableEntity)
		return
	}

	// Create job and enqueue
	job := models.EmailJob{
		To:      req.To,
		Subject: req.Subject,
		Body:    req.Body,
		Retries: 0,
	}

	if err := h.emailService.EnqueueJob(job); err != nil {
		http.Error(w, "Queue is full", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"message": "Email queued for processing",
	})
}

// DeadLetterHandler handles GET /dead-letter requests
func (h *EmailHandler) DeadLetterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs := h.emailService.GetDeadLetterJobs()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": len(jobs),
		"jobs":  jobs,
	})
}

// HealthHandler handles GET /health requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "email-queue",
	})
}
