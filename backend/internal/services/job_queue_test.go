package services

import (
	"testing"
	"time"

	"incident-management-system/internal/database"
	"incident-management-system/internal/models"
	"incident-management-system/internal/storage"

	_ "github.com/mattn/go-sqlite3"
)

func TestJobQueue_NewJobQueue(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	processingService := NewProcessingService(db, fileStore)

	// Test creating job queue with default config
	configQueue := JobQueueConfig{}
	jobQueue := NewJobQueue(configQueue, processingService)
	if jobQueue == nil {
		t.Fatal("Expected non-nil JobQueue")
	}

	if jobQueue.workers != 3 {
		t.Errorf("Expected default 3 workers, got %d", jobQueue.workers)
	}

	// Test creating job queue with custom config
	configQueue = JobQueueConfig{
		Workers:    5,
		BufferSize: 50,
	}
	jobQueue = NewJobQueue(configQueue, processingService)
	if jobQueue == nil {
		t.Fatal("Expected non-nil JobQueue")
	}

	if jobQueue.workers != 5 {
		t.Errorf("Expected 5 workers, got %d", jobQueue.workers)
	}

	// Shutdown the queue
	jobQueue.Shutdown()
}

func TestJobQueue_SubmitJob(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	processingService := NewProcessingService(db, fileStore)

	// Create job queue
	configQueue := JobQueueConfig{
		Workers:    1,
		BufferSize: 10,
	}
	jobQueue := NewJobQueue(configQueue, processingService)
	// Don't defer shutdown here, we'll do it manually

	// Test submitting a job
	payload := map[string]interface{}{
		"test_key": "test_value",
	}
	job, err := jobQueue.SubmitJob(JobTypeProcessUpload, "upload-123", payload)
	if err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}

	if job == nil {
		t.Fatal("Expected non-nil job")
	}

	if job.Type != JobTypeProcessUpload {
		t.Errorf("Expected job type %s, got %s", JobTypeProcessUpload, job.Type)
	}

	if job.UploadID != "upload-123" {
		t.Errorf("Expected upload ID upload-123, got %s", job.UploadID)
	}

	if job.Status != JobStatusPending {
		t.Errorf("Expected status pending, got %s", job.Status)
	}

	// Test submitting job with invalid queue state
	jobQueue.Shutdown() // Shutdown the queue
	_, err = jobQueue.SubmitJob(JobTypeProcessUpload, "upload-456", payload)
	if err == nil {
		t.Error("Expected error when submitting job to shutdown queue")
	}
}

func TestJobQueue_GetJob(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	processingService := NewProcessingService(db, fileStore)

	// Create job queue
	configQueue := JobQueueConfig{
		Workers:    1,
		BufferSize: 10,
	}
	jobQueue := NewJobQueue(configQueue, processingService)
	defer jobQueue.Shutdown()

	// Submit a job
	payload := map[string]interface{}{
		"test_key": "test_value",
	}
	submittedJob, err := jobQueue.SubmitJob(JobTypeProcessUpload, "upload-123", payload)
	if err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}

	// Test retrieving the job
	retrievedJob, err := jobQueue.GetJob(submittedJob.ID)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrievedJob == nil {
		t.Fatal("Expected non-nil job")
	}

	if retrievedJob.ID != submittedJob.ID {
		t.Errorf("Expected job ID %s, got %s", submittedJob.ID, retrievedJob.ID)
	}

	// Test retrieving non-existent job
	_, err = jobQueue.GetJob("non-existent-job")
	if err == nil {
		t.Error("Expected error when getting non-existent job")
	}
}

func TestJobQueue_GetJobsByUpload(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	processingService := NewProcessingService(db, fileStore)

	// Create job queue
	configQueue := JobQueueConfig{
		Workers:    1,
		BufferSize: 10,
	}
	jobQueue := NewJobQueue(configQueue, processingService)
	defer jobQueue.Shutdown()

	// Submit jobs for the same upload
	payload1 := map[string]interface{}{
		"test_key": "test_value_1",
	}
	job1, err := jobQueue.SubmitJob(JobTypeProcessUpload, "upload-123", payload1)
	if err != nil {
		t.Fatalf("Failed to submit job 1: %v", err)
	}

	payload2 := map[string]interface{}{
		"test_key": "test_value_2",
	}
	job2, err := jobQueue.SubmitJob(JobTypeSentimentAnalysis, "upload-123", payload2)
	if err != nil {
		t.Fatalf("Failed to submit job 2: %v", err)
	}

	// Submit job for different upload
	payload3 := map[string]interface{}{
		"test_key": "test_value_3",
	}
	job3, err := jobQueue.SubmitJob(JobTypeAutomationAnalysis, "upload-456", payload3)
	if err != nil {
		t.Fatalf("Failed to submit job 3: %v", err)
	}

	// Test retrieving jobs by upload ID
	jobs := jobQueue.GetJobsByUpload("upload-123")
	if len(jobs) != 2 {
		t.Errorf("Expected 2 jobs for upload-123, got %d", len(jobs))
	}

	// Check that we got the right jobs
	jobIDs := make(map[string]bool)
	for _, job := range jobs {
		jobIDs[job.ID] = true
	}

	if !jobIDs[job1.ID] {
		t.Error("Expected job1 in results")
	}

	if !jobIDs[job2.ID] {
		t.Error("Expected job2 in results")
	}

	if jobIDs[job3.ID] {
		t.Error("Did not expect job3 in results")
	}

	// Test retrieving jobs for non-existent upload
	jobs = jobQueue.GetJobsByUpload("non-existent-upload")
	if len(jobs) != 0 {
		t.Errorf("Expected 0 jobs for non-existent upload, got %d", len(jobs))
	}
}

func TestJobQueue_JobProcessing(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	processingService := NewProcessingService(db, fileStore)

	// Create job queue with fast processing for testing
	configQueue := JobQueueConfig{
		Workers:    1,
		BufferSize: 10,
	}
	jobQueue := NewJobQueue(configQueue, processingService)
	defer jobQueue.Shutdown()

	// Submit a job
	payload := map[string]interface{}{
		"test_key": "test_value",
	}
	job, err := jobQueue.SubmitJob(JobTypeProcessUpload, "upload-123", payload)
	if err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}

	// Give the worker time to process the job
	time.Sleep(100 * time.Millisecond)

	// Check job status (it should have failed since we don't have a real file)
	updatedJob, err := jobQueue.GetJob(job.ID)
	if err != nil {
		t.Fatalf("Failed to get updated job: %v", err)
	}

	// Job should either be failed or retrying due to missing file
	if updatedJob.Status != JobStatusFailed && updatedJob.Status != JobStatusRetrying {
		t.Errorf("Expected job to be failed or retrying, got %s", updatedJob.Status)
	}
}

func TestJobQueue_Shutdown(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	processingService := NewProcessingService(db, fileStore)

	// Create job queue
	configQueue := JobQueueConfig{
		Workers:    2,
		BufferSize: 10,
	}
	jobQueue := NewJobQueue(configQueue, processingService)
	// Don't defer shutdown here, we'll do it manually

	// Submit a few jobs
	for i := 0; i < 5; i++ {
		payload := map[string]interface{}{
			"test_key": "test_value",
		}
		_, err := jobQueue.SubmitJob(JobTypeProcessUpload, "upload-123", payload)
		if err != nil {
			t.Fatalf("Failed to submit job %d: %v", i, err)
		}
	}

	// Shutdown the queue
	jobQueue.Shutdown()

	// Try to submit another job (should fail)
	payload := map[string]interface{}{
		"test_key": "test_value",
	}
	_, err = jobQueue.SubmitJob(JobTypeProcessUpload, "upload-456", payload)
	if err == nil {
		t.Error("Expected error when submitting job after shutdown")
	}
}

func TestJob_GenerateJobID(t *testing.T) {
	// Generate a few job IDs and check they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateJobID()
		if ids[id] {
			t.Errorf("Duplicate job ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestJobQueue_HandleJobError(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	processingService := NewProcessingService(db, fileStore)

	// Create job queue with fast processing for testing
	configQueue := JobQueueConfig{
		Workers:    1,
		BufferSize: 10,
	}
	jobQueue := NewJobQueue(configQueue, processingService)
	// Don't defer shutdown here, we'll do it manually

	// Create a job with 0 max retries to test immediate failure
	job := &Job{
		ID:         "test-job-1",
		Type:       JobTypeProcessUpload,
		Status:     JobStatusPending,
		UploadID:   "upload-123",
		Payload:    map[string]interface{}{},
		MaxRetries: 0,
		CreatedAt:  time.Now(),
	}

	// Store the job
	jobQueue.jobStoreMux.Lock()
	jobQueue.jobStore[job.ID] = job
	jobQueue.jobStoreMux.Unlock()

	// Simulate job error handling
	testErr := &models.ValidationError{
		Field:   "test",
		Value:   "test",
		Message: "test error",
		Row:     1,
	}
	jobQueue.handleJobError(job, testErr)

	// Check that job is marked as failed
	if job.Status != JobStatusFailed {
		t.Errorf("Expected job status failed, got %s", job.Status)
	}

	if job.Error == "" {
		t.Error("Expected error message to be set")
	}

	// Create a job with retries to test retry logic
	job2 := &Job{
		ID:         "test-job-2",
		Type:       JobTypeProcessUpload,
		Status:     JobStatusPending,
		UploadID:   "upload-123",
		Payload:    map[string]interface{}{},
		MaxRetries: 2,
		CreatedAt:  time.Now(),
	}

	// Store the job
	jobQueue.jobStoreMux.Lock()
	jobQueue.jobStore[job2.ID] = job2
	jobQueue.jobStoreMux.Unlock()

	// Simulate job error handling with retries
	jobQueue.handleJobError(job2, testErr)

	// Check that job is marked as retrying
	if job2.Status != JobStatusRetrying {
		t.Errorf("Expected job status retrying, got %s", job2.Status)
	}

	if job2.RetryCount != 1 {
		t.Errorf("Expected retry count 1, got %d", job2.RetryCount)
	}

	// Shutdown the queue
	jobQueue.Shutdown()
}
