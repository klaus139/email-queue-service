# Email Queue Service - Submission Guide

## 📋 Project Overview

This is a complete implementation of the email queue microservice requirements with all bonus features implemented. The service is built using Go with a modular architecture for maintainability and extensibility.

## Quick Start Instructions

### Prerequisites
- Go 1.23.3 or later
- Git

### Running the Service

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd email-queue-service
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run the service:**
   ```bash
   go run .
   ```

4. **Test the service:**
   ```bash
   # Make test script executable
   chmod +x test.sh
   
   # Run comprehensive tests
   ./test.sh
   ```

## 📡 API Testing

### Manual Testing

1. **Health Check:**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Send Valid Email:**
   ```bash
   curl -X POST http://localhost:8080/send-email \
     -H "Content-Type: application/json" \
     -d '{"to": "test@example.com", "subject": "Hello", "body": "Test message"}'
   ```

3. **Test Validation (Invalid Email):**
   ```bash
   curl -X POST http://localhost:8080/send-email \
     -H "Content-Type: application/json" \
     -d '{"to": "invalid-email", "subject": "Test", "body": "Test"}'
   ```

4. **Test Retry Logic:**
   ```bash
   curl -X POST http://localhost:8080/send-email \
     -H "Content-Type: application/json" \
     -d '{"to": "test@example.com", "subject": "This will fail!", "body": "Test"}'
   ```

5. **View Dead Letter Queue:**
   ```bash
   curl http://localhost:8080/dead-letter
   ```

6. **View Metrics:**
   ```bash
   curl http://localhost:8080/metrics
   ```

## ⚙️ Configuration

The service can be configured using environment variables:

```bash
export WORKERS=5        # Number of worker goroutines (default: 3)
export QUEUE_SIZE=200   # Maximum queue size (default: 100)
export PORT=9090        # HTTP server port (default: 8080)
go run .
```

## 🏗️ Architecture

The project follows a clean modular architecture:

- **`main.go`**: Application entry point and HTTP server setup
- **`models/`**: Data structures and types
- **`service/`**: Core business logic and worker management
- **`handlers/`**: HTTP request handlers
- **`config/`**: Configuration management
- **`utils/`**: Utility functions

## Requirements Met

### Core Requirements
- HTTP API with `POST /send-email` endpoint
- Input validation (required fields, email format)
- In-memory job queue with configurable size
- Multiple concurrent workers (configurable, default: 3)
- Graceful shutdown handling SIGINT/SIGTERM
- Proper HTTP status codes (202, 422, 503)

### Bonus Features
- **Retry Logic**: Failed jobs retry up to 3 times with exponential backoff
- **Dead Letter Queue**: Permanently failed jobs stored separately
- **Prometheus Metrics**: Queue length, jobs processed, failures tracked
- **Configurable Workers**: Set workers and queue size via environment variables
- **Health Check Endpoint**: `/health` for monitoring
- **Panic Recovery**: Workers recover from panics gracefully
- **Dead Letter API**: `/dead-letter` endpoint to view failed jobs

## Testing

The project includes a comprehensive test script (`test.sh`) that validates:

- Health endpoint functionality
- Valid email submission
- Input validation (invalid email, missing fields)
- Retry logic with failure simulation
- Dead letter queue access
- Prometheus metrics
- Graceful shutdown

Run the tests with:
```bash
./test.sh
```

## Monitoring

The service exposes Prometheus metrics at `/metrics`:

- `email_queue_length`: Current number of jobs in the queue
- `email_jobs_processed_total`: Total number of processed jobs
- `email_jobs_failed_total`: Total number of permanently failed jobs
- `email_dead_letter_jobs_total`: Total number of jobs in dead letter queue

## 🔧 Development

### Building
```bash
go build -o email-queue-service .
./email-queue-service
```

### Code Quality
- Clean separation of concerns
- Proper error handling
- Comprehensive logging
- Thread-safe operations
- Graceful shutdown implementation

## Notes

- The service simulates email sending with a 1-second delay
- Jobs with subjects ending in '!' are designed to fail for retry testing
- All operations are thread-safe using Go's channel primitives
- The service handles graceful shutdown within 30 seconds
- Prometheus metrics are updated in real-time

## Troubleshooting

If you encounter issues:

1. **Port already in use**: Change the port with `export PORT=8081`
2. **Queue full**: Increase queue size with `export QUEUE_SIZE=500`
3. **High memory usage**: Reduce workers with `export WORKERS=2`

---

**Built with Go 1.23.3** | **All requirements and bonus features implemented** 