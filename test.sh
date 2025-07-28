#!/bin/bash

# Email Queue Service Test Script
# This script tests all functionality of the email queue service

set -e

BASE_URL="http://localhost:8080"
SERVICE_PID=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Start the service
start_service() {
    log_info "Starting email queue service..."
    go run . > service.log 2>&1 &
    SERVICE_PID=$!
    sleep 3
    
    # Check if service started successfully
    if curl -s "$BASE_URL/health" > /dev/null; then
        log_success "Service started successfully (PID: $SERVICE_PID)"
    else
        log_error "Failed to start service"
        exit 1
    fi
}

# Stop the service
stop_service() {
    if [ ! -z "$SERVICE_PID" ]; then
        log_info "Stopping service (PID: $SERVICE_PID)..."
        kill $SERVICE_PID 2>/dev/null || true
        wait $SERVICE_PID 2>/dev/null || true
        log_success "Service stopped"
    fi
}

# Cleanup on exit
cleanup() {
    stop_service
    rm -f service.log
}

trap cleanup EXIT

# Test health endpoint
test_health() {
    log_info "Testing health endpoint..."
    response=$(curl -s "$BASE_URL/health")
    if echo "$response" | grep -q "healthy"; then
        log_success "Health check passed"
        echo "Response: $response"
    else
        log_error "Health check failed"
        return 1
    fi
}

# Test valid email submission
test_valid_email() {
    log_info "Testing valid email submission..."
    response=$(curl -s -X POST "$BASE_URL/send-email" \
        -H "Content-Type: application/json" \
        -d '{"to": "test@example.com", "subject": "Test Email", "body": "This is a test email"}')
    
    if echo "$response" | grep -q "accepted"; then
        log_success "Valid email accepted"
        echo "Response: $response"
    else
        log_error "Valid email submission failed"
        return 1
    fi
}

# Test invalid email format
test_invalid_email() {
    log_info "Testing invalid email format..."
    response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/send-email" \
        -H "Content-Type: application/json" \
        -d '{"to": "invalid-email", "subject": "Test", "body": "Test"}')
    
    http_code="${response: -3}"
    if [ "$http_code" = "422" ]; then
        log_success "Invalid email correctly rejected (422)"
    else
        log_error "Invalid email not properly rejected (got $http_code)"
        return 1
    fi
}

# Test missing fields
test_missing_fields() {
    log_info "Testing missing fields..."
    response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/send-email" \
        -H "Content-Type: application/json" \
        -d '{"to": "test@example.com", "subject": "Test"}')
    
    http_code="${response: -3}"
    if [ "$http_code" = "422" ]; then
        log_success "Missing fields correctly rejected (422)"
    else
        log_error "Missing fields not properly rejected (got $http_code)"
        return 1
    fi
}

# Test retry logic
test_retry_logic() {
    log_info "Testing retry logic (email with subject ending in '!')..."
    response=$(curl -s -X POST "$BASE_URL/send-email" \
        -H "Content-Type: application/json" \
        -d '{"to": "retry@example.com", "subject": "This will fail!", "body": "Testing retry logic"}')
    
    if echo "$response" | grep -q "accepted"; then
        log_success "Retry test email accepted"
        echo "Response: $response"
        log_info "Waiting for retry processing..."
        sleep 10
    else
        log_error "Retry test email submission failed"
        return 1
    fi
}

# Test dead letter queue
test_dead_letter() {
    log_info "Testing dead letter queue..."
    response=$(curl -s "$BASE_URL/dead-letter")
    
    if echo "$response" | grep -q "count"; then
        log_success "Dead letter queue accessible"
        echo "Response: $response"
    else
        log_error "Dead letter queue not accessible"
        return 1
    fi
}

# Test metrics endpoint
test_metrics() {
    log_info "Testing Prometheus metrics..."
    response=$(curl -s "$BASE_URL/metrics")
    
    if echo "$response" | grep -q "email_queue_length"; then
        log_success "Metrics endpoint working"
        echo "Available metrics:"
        echo "$response" | grep "email_" | head -5
    else
        log_error "Metrics endpoint not working"
        return 1
    fi
}

# Test graceful shutdown
test_graceful_shutdown() {
    log_info "Testing graceful shutdown..."
    
    # Send a few emails
    for i in {1..3}; do
        curl -s -X POST "$BASE_URL/send-email" \
            -H "Content-Type: application/json" \
            -d "{\"to\": \"shutdown$i@example.com\", \"subject\": \"Shutdown Test $i\", \"body\": \"Testing shutdown\"}" > /dev/null
    done
    
    # Send SIGTERM
    kill -TERM $SERVICE_PID
    
    # Wait for graceful shutdown
    timeout=30
    while [ $timeout -gt 0 ] && kill -0 $SERVICE_PID 2>/dev/null; do
        sleep 1
        timeout=$((timeout - 1))
    done
    
    if ! kill -0 $SERVICE_PID 2>/dev/null; then
        log_success "Graceful shutdown completed"
    else
        log_warning "Service did not shutdown gracefully within 30 seconds"
        kill -KILL $SERVICE_PID 2>/dev/null || true
    fi
}

# Main test execution
main() {
    log_info "Starting Email Queue Service Tests"
    echo "======================================"
    
    # Start service
    start_service
    
    # Run tests
    test_health
    test_valid_email
    test_invalid_email
    test_missing_fields
    test_retry_logic
    test_dead_letter
    test_metrics
    
    # Test graceful shutdown
    test_graceful_shutdown
    
    echo ""
    log_success "All tests completed successfully!"
    echo "======================================"
}

# Run main function
main "$@"