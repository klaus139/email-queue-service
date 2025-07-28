package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"email-queue-service/models"

	"github.com/prometheus/client_golang/prometheus"
)

// EmailService handles email queue operations
type EmailService struct {
	jobQueue       chan models.EmailJob
	retryQueue     chan models.EmailJob
	deadLetterLog  []models.EmailJob
	workers        int
	queueSize      int
	wg             sync.WaitGroup
	shutdown       chan bool
	deadLetterLock sync.RWMutex

	// Prometheus metrics
	queueLength    prometheus.Gauge
	jobsProcessed  prometheus.Counter
	jobsFailed     prometheus.Counter
	deadLetterJobs prometheus.Counter
}

// NewEmailService creates a new email service
func NewEmailService(workers, queueSize int) *EmailService {
	service := &EmailService{
		jobQueue:      make(chan models.EmailJob, queueSize),
		retryQueue:    make(chan models.EmailJob, queueSize/2), // Smaller retry queue
		deadLetterLog: make([]models.EmailJob, 0),
		workers:       workers,
		queueSize:     queueSize,
		shutdown:      make(chan bool),

		// Initialize Prometheus metrics
		queueLength: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "email_queue_length",
			Help: "Current number of jobs in the email queue",
		}),
		jobsProcessed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "email_jobs_processed_total",
			Help: "Total number of email jobs processed",
		}),
		jobsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "email_jobs_failed_total",
			Help: "Total number of email jobs that failed permanently",
		}),
		deadLetterJobs: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "email_dead_letter_jobs_total",
			Help: "Total number of jobs moved to dead letter queue",
		}),
	}

	// Register metrics
	prometheus.MustRegister(service.queueLength)
	prometheus.MustRegister(service.jobsProcessed)
	prometheus.MustRegister(service.jobsFailed)
	prometheus.MustRegister(service.deadLetterJobs)

	return service
}

// Start initializes workers and monitoring
func (es *EmailService) Start() {
	// Start workers
	for i := 0; i < es.workers; i++ {
		es.wg.Add(1)
		go es.worker(i + 1)
	}

	// Start retry worker
	es.wg.Add(1)
	go es.retryWorker()

	// Start queue length monitoring
	go es.monitorQueueLength()

	log.Printf("Started %d workers with queue size %d", es.workers, es.queueSize)
}

// EnqueueJob adds a job to the queue
func (es *EmailService) EnqueueJob(job models.EmailJob) error {
	select {
	case es.jobQueue <- job:
		return nil
	default:
		return fmt.Errorf("queue is full")
	}
}

// worker processes jobs from the queue
func (es *EmailService) worker(id int) {
	defer es.wg.Done()

	log.Printf("Worker %d started", id)

	for {
		select {
		case job := <-es.jobQueue:
			es.processJob(job, id)
		case job := <-es.retryQueue:
			es.processJob(job, id)
		case <-es.shutdown:
			log.Printf("Worker %d shutting down", id)
			return
		}
	}
}

// retryWorker handles retry logic
func (es *EmailService) retryWorker() {
	defer es.wg.Done()

	log.Println("Retry worker started")

	for {
		select {
		case <-es.shutdown:
			log.Println("Retry worker shutting down")
			return
		default:
			// Process any remaining retry jobs during shutdown
			select {
			case job := <-es.retryQueue:
				es.processJob(job, 0) // 0 indicates retry worker
			case <-time.After(100 * time.Millisecond):
				// Short timeout to check shutdown frequently
			}
		}
	}
}

// processJob simulates sending an email
func (es *EmailService) processJob(job models.EmailJob, workerID int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker %d recovered from panic: %v", workerID, r)
		}
	}()

	log.Printf("Worker %d processing email to %s: %s", workerID, job.To, job.Subject)

	// Simulate email sending with potential failure (10% failure rate for demo)
	time.Sleep(1 * time.Second)

	// Simulate occasional failures for retry demonstration
	if job.Retries == 0 && len(job.Subject) > 10 && job.Subject[len(job.Subject)-1] == '!' {
		// Fail jobs ending with '!' on first try
		es.handleJobFailure(job)
		return
	}

	log.Printf("Worker %d successfully sent email to %s", workerID, job.To)
	es.jobsProcessed.Inc()
}

// handleJobFailure manages retry logic and dead letter queue
func (es *EmailService) handleJobFailure(job models.EmailJob) {
	job.Retries++

	if job.Retries <= 3 {
		log.Printf("Job failed, retrying (%d/3): %s", job.Retries, job.To)

		// Add delay before retry
		go func() {
			time.Sleep(time.Duration(job.Retries) * time.Second)
			select {
			case es.retryQueue <- job:
			default:
				// If retry queue is full, move to dead letter
				es.moveToDeadLetter(job)
			}
		}()
	} else {
		log.Printf("Job permanently failed after 3 retries: %s", job.To)
		es.moveToDeadLetter(job)
	}
}

// moveToDeadLetter adds job to dead letter queue
func (es *EmailService) moveToDeadLetter(job models.EmailJob) {
	es.deadLetterLock.Lock()
	defer es.deadLetterLock.Unlock()

	es.deadLetterLog = append(es.deadLetterLog, job)
	es.jobsFailed.Inc()
	es.deadLetterJobs.Inc()

	log.Printf("Job moved to dead letter queue: %s", job.To)
}

// GetDeadLetterJobs returns copy of dead letter jobs
func (es *EmailService) GetDeadLetterJobs() []models.EmailJob {
	es.deadLetterLock.RLock()
	defer es.deadLetterLock.RUnlock()

	jobs := make([]models.EmailJob, len(es.deadLetterLog))
	copy(jobs, es.deadLetterLog)
	return jobs
}

// monitorQueueLength updates Prometheus gauge
func (es *EmailService) monitorQueueLength() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			es.queueLength.Set(float64(len(es.jobQueue)))
		case <-es.shutdown:
			return
		}
	}
}

// Shutdown gracefully stops the service
func (es *EmailService) Shutdown() {
	log.Println("Shutting down email service...")

	// Close job queue to prevent new jobs
	close(es.jobQueue)

	// Signal all workers to stop
	close(es.shutdown)

	// Wait for all workers to finish
	es.wg.Wait()

	log.Println("Email service shutdown complete")
}
