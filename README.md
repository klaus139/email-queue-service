# Email Queue Service

A robust microservice in Go that accepts email jobs over HTTP, queues them, and processes them asynchronously using a worker system with advanced features like retry logic, dead letter queues, and Prometheus metrics.

##  Features

###  Core Requirements Met

- **HTTP API**: `POST /send-email` endpoint with proper validation
- **Job Queue**: In-memory channel-based queue with multiple concurrent workers
- **Graceful Shutdown**: Handles SIGINT/SIGTERM signals properly
- **Email Validation**: Basic email format validation
- **Error Handling**: Comprehensive error responses and logging

### Bonus Features Implemented

- **Retry Logic**: Failed jobs are retried up to 3 times with exponential backoff
- **Dead Letter Queue**: Permanently failed jobs are stored and accessible via API
- **Prometheus Metrics**: Real-time metrics for monitoring
- **Configurable Workers**: Environment-based configuration
- **Modular Architecture**: Clean separation of concerns with proper Go modules

## API Endpoints

### POST /send-email
Submit an email job for processing.

**Request:**
```json
{
  "to": "user@example.com",
  "subject": "Welcome!",
  "body": "Thanks for signing up."
}
```

**Responses:**
- `202 Accepted`: Email queued successfully
- `422 Bad Request`: Invalid input (missing fields or invalid email)
- `503 Service Unavailable`: Queue is full

### GET /dead-letter
Retrieve failed jobs from the dead letter queue.

**Response:**
```json
{
  "count": 2,
  "jobs": [
    {
      "to": "user@example.com",
      "subject": "Failed Email",
      "body": "This email failed permanently"
    }
  ]
}
```

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "service": "email-queue"
}
```

### GET /metrics
Prometheus metrics endpoint.

## Architecture

The service is built with a modular architecture:

```
email-queue-service/
├── main.go              # Application entry point
├── models/
│   └── email.go         # Data structures
├── service/
│   └── email_service.go # Core business logic
├── handlers/
│   └── http_handlers.go # HTTP request handlers
├── config/
│   └── config.go        # Configuration management
├── utils/
│   └── validation.go    # Validation utilities
└── go.mod               # Go module definition
```

## Quick Start

### Prerequisites
- Go 1.23.3 or later

### Running the Service

1. **Clone and navigate to the project:**
   ```bash
   cd email-queue-service
   ```

2. **Run the service:**
   ```bash
   go run .
   ```

3. **Test the API:**
   ```bash
   # Send an email
   curl -X POST http://localhost:8080/send-email \
     -H "Content-Type: application/json" \
     -d '{"to": "test@example.com", "subject": "Hello", "body": "Test message"}'
   
   # Check health
   curl http://localhost:8080/health
   
   # View metrics
   curl http://localhost:8080/metrics
   ```



## Configuration

The service can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `WORKERS` | 3 | Number of worker goroutines |
| `QUEUE_SIZE` | 100 | Maximum size of the job queue |
| `PORT` | 8080 | HTTP server port |

Example:
```bash
export WORKERS=5
export QUEUE_SIZE=200
export PORT=9090
go run .
```

## Monitoring

### Prometheus Metrics

The service exposes the following metrics:

- `email_queue_length`: Current number of jobs in the queue
- `email_jobs_processed_total`: Total number of processed jobs
- `email_jobs_failed_total`: Total number of permanently failed jobs
- `email_dead_letter_jobs_total`: Total number of jobs in dead letter queue

### Example Prometheus Query
```promql
# Queue utilization
rate(email_jobs_processed_total[5m])

# Failed job rate
rate(email_jobs_failed_total[5m])
```

## Retry Logic

The service implements intelligent retry logic:

1. **First Failure**: Job is retried after 1 second
2. **Second Failure**: Job is retried after 2 seconds  
3. **Third Failure**: Job is retried after 3 seconds
4. **Final Failure**: Job is moved to dead letter queue

### Testing Retry Logic

To test retry functionality, send an email with a subject ending in `!`:

```bash
curl -X POST http://localhost:8080/send-email \
  -H "Content-Type: application/json" \
  -d '{"to": "test@example.com", "subject": "This will fail!", "body": "Test"}'
```

## Error Handling

The service includes comprehensive error handling:

- **Panic Recovery**: Workers recover from panics automatically
- **Graceful Shutdown**: Proper cleanup on termination signals
- **Queue Overflow**: Handles queue full scenarios
- **Invalid Input**: Validates all incoming requests

## Testing

### Manual Testing

1. **Basic Functionality:**
   ```bash
   # Send valid email
   curl -X POST http://localhost:8080/send-email \
     -H "Content-Type: application/json" \
     -d '{"to": "test@example.com", "subject": "Test", "body": "Test"}'
   ```

2. **Validation Testing:**
   ```bash
   # Test invalid email
   curl -X POST http://localhost:8080/send-email \
     -H "Content-Type: application/json" \
     -d '{"to": "invalid-email", "subject": "Test", "body": "Test"}'
   
   # Test missing fields
   curl -X POST http://localhost:8080/send-email \
     -H "Content-Type: application/json" \
     -d '{"to": "test@example.com", "subject": "Test"}'
   ```

3. **Retry Testing:**
   ```bash
   # Send email that will fail and retry
   curl -X POST http://localhost:8080/send-email \
     -H "Content-Type: application/json" \
     -d '{"to": "test@example.com", "subject": "This will fail!", "body": "Test"}'
   ```

## Performance

- **Concurrent Processing**: Multiple workers process jobs simultaneously
- **Non-blocking Queue**: Channel-based queue with configurable size
- **Efficient Memory Usage**: Minimal memory footprint
- **Fast Response Times**: Sub-millisecond API response times

## Development

### Project Structure
```
├── main.go              # Application entry point and HTTP server setup
├── models/              # Data models and structures
├── service/             # Core business logic and worker management
├── handlers/            # HTTP request handlers
├── config/              # Configuration management
├── utils/               # Utility functions
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── test.sh              # Comprehensive test script
└── README.md            # This file
```

### Adding New Features

The modular architecture makes it easy to extend:

1. **New Endpoints**: Add handlers in `handlers/`
2. **New Models**: Add structures in `models/`
3. **New Services**: Add business logic in `service/`
4. **New Utilities**: Add helper functions in `utils/`

## Troubleshooting

### Common Issues

1. **Port Already in Use:**
   ```bash
   # Change port
   export PORT=8081
   go run .
   ```

2. **Queue Full:**
   ```bash
   # Increase queue size
   export QUEUE_SIZE=500
   go run .
   ```

3. **High Memory Usage:**
   ```bash
   # Reduce workers
   export WORKERS=2
   go run .
   ```

## 📝 License

This project is open source and available under the MIT License.

##  Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

---
