package services

import "incident-management-system/internal/models"

// SentimentResult represents the result of sentiment analysis
type SentimentResult struct {
	Score float64 `json:"score"` // -1.0 to 1.0
	Label string  `json:"label"` // positive, negative, neutral
}

// SentimentAnalyzer interface for sentiment analysis services
type SentimentAnalyzer interface {
	AnalyzeSentiment(text string) (*SentimentResult, error)
	AnalyzeBatch(texts []string) ([]*SentimentResult, error)
}

// AutomationResult represents the result of automation analysis
type AutomationResult struct {
	Score          float64 `json:"score"`           // 0.0 to 1.0
	Feasible       bool    `json:"feasible"`        // true if automation is recommended
	ITProcessGroup string  `json:"it_process_group"` // categorized IT process group
	Confidence     float64 `json:"confidence"`      // confidence in the analysis
	Reasons        []string `json:"reasons"`        // reasons for the score
}

// AutomationAnalyzer interface for automation analysis services
type AutomationAnalyzer interface {
	AnalyzeAutomation(incident *models.Incident) (*AutomationResult, error)
	AnalyzeBatch(incidents []*models.Incident) ([]*AutomationResult, error)
}

// ProcessingEngine interface for coordinating data processing
type ProcessingEngine interface {
	ProcessUpload(uploadID string) error
	GetProcessingStatus(uploadID string) (*ProcessingProgress, error)
	SubmitAnalysisJob(uploadID string, analysisType string) (string, error)
	GetJobStatus(jobID string) (*Job, error)
}