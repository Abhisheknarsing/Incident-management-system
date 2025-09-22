package services

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"incident-management-system/internal/models"

	"github.com/xuri/excelize/v2"
)

// ExcelParser handles parsing of Excel files with concurrent processing
type ExcelParser struct {
	maxWorkers int
	batchSize  int
}

// ExcelParserConfig holds configuration for the Excel parser
type ExcelParserConfig struct {
	MaxWorkers int // Maximum number of concurrent workers
	BatchSize  int // Number of rows to process in each batch
}

// DefaultExcelParserConfig returns default configuration
func DefaultExcelParserConfig() *ExcelParserConfig {
	return &ExcelParserConfig{
		MaxWorkers: runtime.NumCPU(), // Use all available CPU cores
		BatchSize:  100,              // Process 100 rows per batch
	}
}

// NewExcelParser creates a new Excel parser with the given configuration
func NewExcelParser(config *ExcelParserConfig) *ExcelParser {
	if config == nil {
		config = DefaultExcelParserConfig()
	}

	// Ensure reasonable limits
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = 1
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 10
	}

	return &ExcelParser{
		maxWorkers: config.MaxWorkers,
		batchSize:  config.BatchSize,
	}
}

// ParseFile parses an Excel file and returns incidents with concurrent processing
func (p *ExcelParser) ParseFile(ctx context.Context, filePath string) ([]models.Incident, error) {
	// Open Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get all rows from the first sheet
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		// Try to get rows from the first available sheet
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("no sheets found in Excel file")
		}
		rows, err = f.GetRows(sheets[0])
		if err != nil {
			return nil, fmt.Errorf("failed to read rows from sheet: %w", err)
		}
	}

	// Check if we have data
	if len(rows) <= 1 {
		return []models.Incident{}, nil
	}

	// Parse header row to get column indices
	header := rows[0]
	columnIndices := p.parseHeader(header)

	// Process data rows concurrently
	dataRows := rows[1:]
	incidents, err := p.processRowsConcurrently(ctx, dataRows, columnIndices)
	if err != nil {
		return nil, fmt.Errorf("failed to process rows: %w", err)
	}

	return incidents, nil
}

// parseHeader maps column names to indices
func (p *ExcelParser) parseHeader(header []string) map[string]int {
	indices := make(map[string]int)

	// Define expected column names and their mappings
	columnMappings := map[string][]string{
		"incident_id":         {"incidentid", "incidentid", "id", "ticketid", "ticketid"},
		"application_name":    {"applicationname", "applicationname", "app", "application"},
		"report_date":         {"reportdate", "reportdate", "date", "createddate", "createddate"},
		"priority":            {"priority", "prio", "severity"},
		"status":              {"status", "state"},
		"resolved_person":     {"resolvedperson", "resolver", "resolvedby", "resolvedby"},
		"resolve_date":        {"resolvedate", "resolvedate", "resolveddate", "resolveddate"},
		"brief_description":   {"briefdescription", "description", "desc", "summary"},
		"resolution_group":    {"resolutiongroup", "assignee", "assignedto", "assignedto"},
		"it_process_group":    {"itprocessgroup", "itprocessgroup", "processgroup", "processgroup"},
		"automation_feasible": {"automationfeasible", "automationfeasible", "automatable"},
		"automation_score":    {"automationscore", "automationscore"},
		"sentiment_label":     {"sentimentlabel", "sentimentlabel", "sentiment"},
		"sentiment_score":     {"sentimentscore", "sentimentscore"},
		"closure_code":        {"closurecode", "closurecode", "closecode", "closecode"},
	}

	// Map header columns to expected fields
	for i, columnName := range header {
		// Normalize column name (lowercase, remove spaces)
		normalized := normalizeColumnName(columnName)

		// Find matching field
		for field, possibleNames := range columnMappings {
			for _, possibleName := range possibleNames {
				if normalized == possibleName {
					indices[field] = i
					break
				}
			}
		}
	}

	return indices
}

// normalizeColumnName normalizes column names for matching
func normalizeColumnName(name string) string {
	// Convert to lowercase and remove spaces, underscores, hyphens
	result := ""
	for _, r := range name {
		if r != ' ' && r != '_' && r != '-' {
			result += string(r)
		}
	}
	return strings.ToLower(result)
}

// processRowsConcurrently processes rows using concurrent workers
func (p *ExcelParser) processRowsConcurrently(ctx context.Context, rows [][]string, columnIndices map[string]int) ([]models.Incident, error) {
	// Create channels for work distribution and results collection
	type workItem struct {
		index int
		row   []string
	}

	workChan := make(chan workItem, len(rows))
	resultsChan := make(chan struct {
		index    int
		incident models.Incident
		err      error
	}, len(rows))

	// Create worker pool
	var wg sync.WaitGroup
	workerCount := p.maxWorkers
	if workerCount > len(rows) {
		workerCount = len(rows)
	}

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case work, ok := <-workChan:
					if !ok {
						return // Channel closed
					}

					// Process the row
					incident, err := p.parseRow(work.row, columnIndices)
					resultsChan <- struct {
						index    int
						incident models.Incident
						err      error
					}{work.index, incident, err}

				case <-ctx.Done():
					return // Context cancelled
				}
			}
		}()
	}

	// Send work to workers
	go func() {
		defer close(workChan)
		for i, row := range rows {
			select {
			case workChan <- workItem{index: i, row: row}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	incidents := make([]models.Incident, len(rows))
	var firstError error

	for result := range resultsChan {
		if result.err != nil {
			if firstError == nil {
				firstError = result.err
			}
			continue
		}
		incidents[result.index] = result.incident
	}

	// Filter out zero-value incidents (failed parses)
	filtered := make([]models.Incident, 0, len(incidents))
	for _, incident := range incidents {
		if incident.IncidentID != "" {
			filtered = append(filtered, incident)
		}
	}

	if firstError != nil {
		return filtered, firstError
	}

	return filtered, nil
}

// parseRow parses a single row into an Incident model
func (p *ExcelParser) parseRow(row []string, columnIndices map[string]int) (models.Incident, error) {
	incident := models.Incident{}
	incident.SetDefaults()

	// Helper function to safely get cell value
	getCellValue := func(fieldName string) string {
		if index, exists := columnIndices[fieldName]; exists && index < len(row) {
			return row[index]
		}
		return ""
	}

	// Parse required fields
	incident.IncidentID = getCellValue("incident_id")
	if incident.IncidentID == "" {
		return models.Incident{}, fmt.Errorf("missing required field: incident_id")
	}

	incident.ApplicationName = getCellValue("application_name")
	incident.BriefDescription = getCellValue("brief_description")
	incident.ResolutionGroup = getCellValue("resolution_group")
	incident.ResolvedPerson = getCellValue("resolved_person")
	incident.Priority = getCellValue("priority")
	incident.Status = getCellValue("status")
	incident.ITProcessGroup = getCellValue("it_process_group")
	incident.SentimentLabel = getCellValue("sentiment_label")

	// Parse date fields
	if dateStr := getCellValue("report_date"); dateStr != "" {
		if parsedDate, err := parseDate(dateStr); err == nil {
			incident.ReportDate = parsedDate
		}
	}

	if dateStr := getCellValue("resolve_date"); dateStr != "" {
		if parsedDate, err := parseDate(dateStr); err == nil {
			incident.ResolveDate = &parsedDate
		}
	}

	// Parse numeric fields
	if scoreStr := getCellValue("automation_score"); scoreStr != "" {
		if score, err := strconv.ParseFloat(scoreStr, 64); err == nil {
			incident.AutomationScore = &score
		}
	}

	if scoreStr := getCellValue("sentiment_score"); scoreStr != "" {
		if score, err := strconv.ParseFloat(scoreStr, 64); err == nil {
			incident.SentimentScore = &score
		}
	}

	// Parse boolean fields
	if feasibleStr := getCellValue("automation_feasible"); feasibleStr != "" {
		feasible := feasibleStr == "true" || feasibleStr == "1" || feasibleStr == "yes"
		incident.AutomationFeasible = &feasible
	}

	return incident, nil
}

// parseDate attempts to parse a date string in various formats
func parseDate(dateStr string) (time.Time, error) {
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"2006/01/02",
		"01-02-2006",
		"02-01-2006",
		"2006-01-02 15:04:05",
		"01/02/2006 15:04:05",
		time.RFC3339,
		time.RFC822,
	}

	for _, format := range formats {
		if parsedDate, err := time.Parse(format, dateStr); err == nil {
			return parsedDate, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
