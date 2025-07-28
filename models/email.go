package models

// EmailJob represents an email to be sent
type EmailJob struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Retries int    `json:"-"`
}

// EmailRequest represents the incoming HTTP request
type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
