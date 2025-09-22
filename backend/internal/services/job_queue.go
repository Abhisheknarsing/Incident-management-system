package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"incident-management-system/internal/models"
)

// JobType represents the type of job to be processed
type JobType string

const (
	JobTypeProcessUpload      JobType = "process_upload"
	JobTypeSentimentAnalysis  JobType = "sentiment_analysis"
	JobTypeAutomationAnalysis JobType = "automation_analysis"
)

// JobStatus represents the current status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRetrying  JobStatus = "retrying"
)

// Job represents a processing job in the queue
type Job struct {
	ID          string                 `json:"id"`
	Type        JobType                `json:"type"`
	Status      JobStatus              `json:"status"`
	UploadID    string                 `json:"upload_id"`
	Payload     map[string]interface{} `json:"payload"`
	Progress    int                    `json:"progress"` // 0-100
	Message     string                 `json:"message"`
	Error       string                 `json:"error,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Result      interface{}            `json:"result,omitempty"`
}

// JobQueue manages asynchronous job processing
type JobQueue struct {
	jobs        chan *Job
	workers     int
	jobStore    map[string]*Job
	jobStoreMux sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup

	// Services for job processing
	processingService *ProcessingService
	sentimentService  SentimentAnalyzer
	automationService AutomationAnalyzer
}

// JobQueueConfig holds configuration for the job queue
type JobQueueConfig struct {
	Workers    int
	BufferSize int
}

// NewJobQueue creates a new job queue instance
func NewJobQueue(config JobQueueConfig, processingService *ProcessingService) *JobQueue {
	ctx, cancel := context.WithCancel(context.Background())

	if config.Workers <= 0 {
		config.Workers = 3 // Default to 3 workers
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 100 // Default buffer size
	}

	jq := &JobQueue{
		jobs:              make(chan *Job, config.BufferSize),
		workers:           config.Workers,
		jobStore:          make(map[string]*Job),
		ctx:               ctx,
		cancel:            cancel,
		processingService: processingService,
	}

	// Start workers
	jq.startWorkers()

	return jq
}

// SetSentimentService sets the sentiment analysis service
func (jq *JobQueue) SetSentimentService(service SentimentAnalyzer) {
	jq.sentimentService = service
}

// SetAutomationService sets the automation analysis service
func (jq *JobQueue) SetAutomationService(service AutomationAnalyzer) {
	jq.automationService = service
}

// SubmitJob submits a new job to the queue
func (jq *JobQueue) SubmitJob(jobType JobType, uploadID string, payload map[string]interface{}) (*Job, error) {
	job := &Job{
		ID:         generateJobID(),
		Type:       jobType,
		Status:     JobStatusPending,
		UploadID:   uploadID,
		Payload:    payload,
		Progress:   0,
		MaxRetries: 3, // Default max retries
		CreatedAt:  time.Now(),
	}

	// Store job
	jq.jobStoreMux.Lock()
	jq.jobStore[job.ID] = job
	jq.jobStoreMux.Unlock()

	// Submit to queue
	select {
	case jq.jobs <- job:
		log.Printf("Job %s (%s) submitted for upload %s", job.ID, job.Type, uploadID)
		return job, nil
	case <-jq.ctx.Done():
		return nil, fmt.Errorf("job queue is shutting down")
	default:
		return nil, fmt.Errorf("job queue is full")
	}
}

// GetJob retrieves a job by ID
func (jq *JobQueue) GetJob(jobID string) (*Job, error) {
	jq.jobStoreMux.RLock()
	defer jq.jobStoreMux.RUnlock()

	job, exists := jq.jobStore[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// GetJobsByUpload retrieves all jobs for a specific upload
func (jq *JobQueue) GetJobsByUpload(uploadID string) []*Job {
	jq.jobStoreMux.RLock()
	defer jq.jobStoreMux.RUnlock()

	var jobs []*Job
	for _, job := range jq.jobStore {
		if job.UploadID == uploadID {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

// startWorkers starts the worker goroutines
func (jq *JobQueue) startWorkers() {
	for i := 0; i < jq.workers; i++ {
		jq.wg.Add(1)
		go jq.worker(i)
	}
	log.Printf("Started %d job queue workers", jq.workers)
}

// worker processes jobs from the queue
func (jq *JobQueue) worker(workerID int) {
	defer jq.wg.Done()

	log.Printf("Worker %d started", workerID)

	for {
		select {
		case job := <-jq.jobs:
			// Check if job is nil (shouldn't happen but let's be safe)
			if job == nil {
				continue
			}
			jq.processJob(workerID, job)
		case <-jq.ctx.Done():
			log.Printf("Worker %d shutting down", workerID)
			return
		}
	}
}

// processJob processes a single job
func (jq *JobQueue) processJob(workerID int, job *Job) {
	// Safety check for nil job
	if job == nil {
		log.Printf("Worker %d received nil job, skipping", workerID)
		return
	}

	log.Printf("Worker %d processing job %s (%s) for upload %s",
		workerID, job.ID, job.Type, job.UploadID)

	// Update job status to running
	jq.updateJobStatus(job, JobStatusRunning, 0, "Processing started")

	startTime := time.Now()
	job.StartedAt = &startTime

	var err error

	// Process based on job type
	switch job.Type {
	case JobTypeProcessUpload:
		// Check if processing service is available
		if jq.processingService == nil {
			err = fmt.Errorf("processing service not available")
			break
		}
		err = jq.processUploadJob(job)
	case JobTypeSentimentAnalysis:
		// Check if sentiment service is available
		if jq.sentimentService == nil {
			err = fmt.Errorf("sentiment analysis service not available")
			break
		}
		err = jq.processSentimentAnalysisJob(job)
	case JobTypeAutomationAnalysis:
		// Check if automation service is available
		if jq.automationService == nil {
			err = fmt.Errorf("automation analysis service not available")
			break
		}
		err = jq.processAutomationAnalysisJob(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.Type)
	}

	// Handle job completion or failure
	if err != nil {
		jq.handleJobError(job, err)
	} else {
		jq.completeJob(job)
	}
}

// processUploadJob processes an upload job
func (jq *JobQueue) processUploadJob(job *Job) error {
	if jq.processingService == nil {
		return fmt.Errorf("processing service not available")
	}

	// Update progress
	jq.updateJobStatus(job, JobStatusRunning, 10, "Starting file processing")

	// Process the upload
	result, err := jq.processingService.ProcessUpload(jq.ctx, job.UploadID)
	if err != nil {
		return fmt.Errorf("failed to process upload: %w", err)
	}

	// Update progress
	jq.updateJobStatus(job, JobStatusRunning, 90, "File processing completed")

	// Store result
	job.Result = result

	return nil
}

// processSentimentAnalysisJob processes sentiment analysis for incidents
func (jq *JobQueue) processSentimentAnalysisJob(job *Job) error {
	if jq.sentimentService == nil {
		return fmt.Errorf("sentiment analysis service not available")
	}

	// Update progress
	jq.updateJobStatus(job, JobStatusRunning, 10, "Starting sentiment analysis")

	// Get incidents for the upload
	incidents, err := jq.processingService.incidentService.GetIncidentsByUpload(jq.ctx, job.UploadID)
	if err != nil {
		return fmt.Errorf("failed to get incidents: %w", err)
	}

	if len(incidents) == 0 {
		jq.updateJobStatus(job, JobStatusRunning, 100, "No incidents to analyze")
		return nil
	}

	// Process sentiment analysis in batches
	batchSize := 10
	totalIncidents := len(incidents)
	processedCount := 0

	for i := 0; i < len(incidents); i += batchSize {
		end := i + batchSize
		if end > len(incidents) {
			end = len(incidents)
		}

		batch := incidents[i:end]

		// Analyze sentiment for batch
		for j := range batch {
			result, err := jq.sentimentService.AnalyzeSentiment(batch[j].Description)
			if err != nil {
				log.Printf("Warning: Failed to analyze sentiment for incident %s: %v",
					batch[j].IncidentID, err)
				continue
			}

			// Update incident with sentiment data
			batch[j].SentimentScore = &result.Score
			batch[j].SentimentLabel = result.Label
		}

		// Update incidents in database
		err = jq.updateIncidentsSentiment(batch)
		if err != nil {
			return fmt.Errorf("failed to update sentiment data: %w", err)
		}

		processedCount += len(batch)
		progress := int(float64(processedCount)/float64(totalIncidents)*90) + 10
		jq.updateJobStatus(job, JobStatusRunning, progress,
			fmt.Sprintf("Processed sentiment for %d/%d incidents", processedCount, totalIncidents))
	}

	job.Result = map[string]interface{}{
		"processed_incidents": processedCount,
		"total_incidents":     totalIncidents,
	}

	return nil
}

// processAutomationAnalysisJob processes automation analysis for incidents
func (jq *JobQueue) processAutomationAnalysisJob(job *Job) error {
	if jq.automationService == nil {
		return fmt.Errorf("automation analysis service not available")
	}

	// Update progress
	jq.updateJobStatus(job, JobStatusRunning, 10, "Starting automation analysis")

	// Get incidents for the upload
	incidents, err := jq.processingService.incidentService.GetIncidentsByUpload(jq.ctx, job.UploadID)
	if err != nil {
		return fmt.Errorf("failed to get incidents: %w", err)
	}

	if len(incidents) == 0 {
		jq.updateJobStatus(job, JobStatusRunning, 100, "No incidents to analyze")
		return nil
	}

	// Process automation analysis in batches
	batchSize := 10
	totalIncidents := len(incidents)
	processedCount := 0

	for i := 0; i < len(incidents); i += batchSize {
		end := i + batchSize
		if end > len(incidents) {
			end = len(incidents)
		}

		batch := incidents[i:end]

		// Analyze automation potential for batch
		for j := range batch {
			result, err := jq.automationService.AnalyzeAutomation(&batch[j])
			if err != nil {
				log.Printf("Warning: Failed to analyze automation for incident %s: %v",
					batch[j].IncidentID, err)
				continue
			}

			// Update incident with automation data
			batch[j].AutomationScore = &result.Score
			batch[j].AutomationFeasible = &result.Feasible
			batch[j].ITProcessGroup = result.ITProcessGroup
		}

		// Update incidents in database
		err = jq.updateIncidentsAutomation(batch)
		if err != nil {
			return fmt.Errorf("failed to update automation data: %w", err)
		}

		processedCount += len(batch)
		progress := int(float64(processedCount)/float64(totalIncidents)*90) + 10
		jq.updateJobStatus(job, JobStatusRunning, progress,
			fmt.Sprintf("Processed automation analysis for %d/%d incidents", processedCount, totalIncidents))
	}

	job.Result = map[string]interface{}{
		"processed_incidents": processedCount,
		"total_incidents":     totalIncidents,
	}

	return nil
}

// updateJobStatus updates the status and progress of a job
func (jq *JobQueue) updateJobStatus(job *Job, status JobStatus, progress int, message string) {
	jq.jobStoreMux.Lock()
	defer jq.jobStoreMux.Unlock()

	job.Status = status
	job.Progress = progress
	job.Message = message

	log.Printf("Job %s status updated: %s (%d%%) - %s", job.ID, status, progress, message)
}

// completeJob marks a job as completed
func (jq *JobQueue) completeJob(job *Job) {
	completedAt := time.Now()
	job.CompletedAt = &completedAt

	jq.updateJobStatus(job, JobStatusCompleted, 100, "Job completed successfully")

	log.Printf("Job %s completed successfully for upload %s", job.ID, job.UploadID)
}

// handleJobError handles job errors and implements retry logic
func (jq *JobQueue) handleJobError(job *Job, err error) {
	job.Error = err.Error()

	log.Printf("Job %s failed: %v (retry %d/%d)", job.ID, err, job.RetryCount, job.MaxRetries)

	// Check if we should retry
	if job.RetryCount < job.MaxRetries {
		job.RetryCount++
		jq.updateJobStatus(job, JobStatusRetrying, job.Progress,
			fmt.Sprintf("Retrying job (attempt %d/%d): %v", job.RetryCount, job.MaxRetries, err))

		// Exponential backoff for retry
		retryDelay := time.Duration(job.RetryCount*job.RetryCount) * time.Second

		go func() {
			time.Sleep(retryDelay)

			// Reset job for retry
			job.Status = JobStatusPending
			job.Error = ""

			// Resubmit to queue
			select {
			case jq.jobs <- job:
				log.Printf("Job %s resubmitted for retry %d", job.ID, job.RetryCount)
			case <-jq.ctx.Done():
				log.Printf("Cannot retry job %s: queue shutting down", job.ID)
			default:
				log.Printf("Cannot retry job %s: queue is full", job.ID)
				jq.updateJobStatus(job, JobStatusFailed, job.Progress, "Failed to retry: queue full")
			}
		}()
	} else {
		// Max retries exceeded
		completedAt := time.Now()
		job.CompletedAt = &completedAt
		jq.updateJobStatus(job, JobStatusFailed, job.Progress,
			fmt.Sprintf("Job failed after %d retries: %v", job.MaxRetries, err))
	}
}

// Shutdown gracefully shuts down the job queue
func (jq *JobQueue) Shutdown() {
	log.Println("Shutting down job queue...")

	jq.cancel()

	// Close the jobs channel to signal workers to stop
	close(jq.jobs)

	// Wait for all workers to finish
	jq.wg.Wait()

	log.Println("Job queue shutdown complete")
}

// Helper functions

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

// updateIncidentsSentiment updates sentiment data for incidents in the database
func (jq *JobQueue) updateIncidentsSentiment(incidents []models.Incident) error {
	// This would typically be implemented in the incident service
	// For now, we'll implement a simple batch update

	for _, incident := range incidents {
		query := `
			UPDATE incidents 
			SET sentiment_score = ?, sentiment_label = ?, updated_at = ?
			WHERE id = ?
		`

		_, err := jq.processingService.db.ExecContext(jq.ctx, query,
			incident.SentimentScore, incident.SentimentLabel, time.Now(), incident.ID)
		if err != nil {
			return fmt.Errorf("failed to update sentiment for incident %s: %w", incident.ID, err)
		}
	}

	return nil
}

// updateIncidentsAutomation updates automation data for incidents in the database
func (jq *JobQueue) updateIncidentsAutomation(incidents []models.Incident) error {
	// This would typically be implemented in the incident service
	// For now, we'll implement a simple batch update

	for _, incident := range incidents {
		query := `
			UPDATE incidents 
			SET automation_score = ?, automation_feasible = ?, it_process_group = ?, updated_at = ?
			WHERE id = ?
		`

		_, err := jq.processingService.db.ExecContext(jq.ctx, query,
			incident.AutomationScore, incident.AutomationFeasible, incident.ITProcessGroup,
			time.Now(), incident.ID)
		if err != nil {
			return fmt.Errorf("failed to update automation data for incident %s: %w", incident.ID, err)
		}
	}

	return nil
}
