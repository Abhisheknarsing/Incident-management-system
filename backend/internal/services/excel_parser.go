package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"incident-management-system/internal/models"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// ExcelParser handles Excel file parsing and validation
type ExcelParser struct {
	// Column mappings for flexible Excel formats
	columnMappings map[string]string
}

// NewExcelParser creates a new ExcelParser instance
func NewExcelParser() *ExcelParser {
	return &ExcelParser{
		columnMappings: getDefaultColumnMappings(),
	}
}

// ParseResult represents the result of parsing an Excel file
type ParseResult struct {
	Incidents []models.Incident      `json:"incidents"`
	Errors    []models.ValidationError `json:"errors"`
	TotalRows int                    `json:"total_rows"`
	ValidRows int                    `json:"valid_rows"`
}

// ParseExcelFile parses an Excel file and returns incidents with validation errors
func (p *ExcelParser) ParseExcelFile(filePath, uploadID string) (*ParseResult, error) {
	// Open Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first worksheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("no worksheets found in Excel file")
	}

	// Get all rows from the worksheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from worksheet: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("worksheet is empty")
	}

	// Parse header row to determine column positions
	headerRow := rows[0]
	columnMap, err := p.mapColumns(headerRow)
	if err != nil {
		return nil, fmt.Errorf("failed to map columns: %w", err)
	}

	// Validate required columns are present
	if err := p.validateRequiredColumns(columnMap); err != nil {
		return nil, err
	}

	result := &ParseResult{
		Incidents: make([]models.Incident, 0),
		Errors:    make([]models.ValidationError, 0),
		TotalRows: len(rows) - 1, // Exclude header row
		ValidRows: 0,
	}

	// Process data rows
	for rowIndex, row := range rows[1:] { // Skip header row
		rowNumber := rowIndex + 2 // Excel row numbers start at 1, plus header

		incident, validationErrors := p.parseRow(row, columnMap, uploadID, rowNumber)
		
		if len(validationErrors) > 0 {
			result.Errors = append(result.Errors, validationErrors...)
		} else {
			result.Incidents = append(result.Incidents, *incident)
			result.ValidRows++
		}
	}

	return result, nil
}

// mapColumns maps Excel column headers to internal field names
func (p *ExcelParser) mapColumns(headerRow []string) (map[string]int, error) {
	columnMap := make(map[string]int)
	
	for i, header := range headerRow {
		normalizedHeader := p.normalizeColumnName(header)
		
		// Check if this header matches any of our expected columns
		for fieldName, expectedHeader := range p.columnMappings {
			if normalizedHeader == p.normalizeColumnName(expectedHeader) {
				columnMap[fieldName] = i
				break
			}
		}
	}

	return columnMap, nil
}

// normalizeColumnName normalizes column names for comparison
func (p *ExcelParser) normalizeColumnName(name string) string {
	// Convert to lowercase and remove spaces, underscores, and special characters
	normalized := strings.ToLower(name)
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	return normalized
}

// validateRequiredColumns ensures all required columns are present
func (p *ExcelParser) validateRequiredColumns(columnMap map[string]int) error {
	requiredFields := []string{
		"incident_id", "report_date", "brief_description", 
		"application_name", "resolution_group", "resolved_person", "priority",
	}

	var missingFields []string
	for _, field := range requiredFields {
		if _, exists := columnMap[field]; !exists {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required columns: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

// parseRow parses a single Excel row into an Incident
func (p *ExcelParser) parseRow(row []string, columnMap map[string]int, uploadID string, rowNumber int) (*models.Incident, []models.ValidationError) {
	incident := &models.Incident{
		ID:       uuid.New().String(),
		UploadID: uploadID,
	}

	var errors []models.ValidationError

	// Parse required fields
	if col, exists := columnMap["incident_id"]; exists && col < len(row) {
		incident.IncidentID = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["brief_description"]; exists && col < len(row) {
		incident.BriefDescription = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["application_name"]; exists && col < len(row) {
		incident.ApplicationName = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["resolution_group"]; exists && col < len(row) {
		incident.ResolutionGroup = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["resolved_person"]; exists && col < len(row) {
		incident.ResolvedPerson = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["priority"]; exists && col < len(row) {
		incident.Priority = strings.TrimSpace(row[col])
	}

	// Parse dates
	if col, exists := columnMap["report_date"]; exists && col < len(row) {
		if date, err := p.parseDate(row[col]); err != nil {
			errors = append(errors, models.ValidationError{
				Field:   "report_date",
				Value:   row[col],
				Message: fmt.Sprintf("invalid date format: %v", err),
				Row:     rowNumber,
			})
		} else {
			incident.ReportDate = date
		}
	}

	if col, exists := columnMap["resolve_date"]; exists && col < len(row) && strings.TrimSpace(row[col]) != "" {
		if date, err := p.parseDate(row[col]); err != nil {
			errors = append(errors, models.ValidationError{
				Field:   "resolve_date",
				Value:   row[col],
				Message: fmt.Sprintf("invalid date format: %v", err),
				Row:     rowNumber,
			})
		} else {
			incident.ResolveDate = &date
		}
	}

	if col, exists := columnMap["last_resolve_date"]; exists && col < len(row) && strings.TrimSpace(row[col]) != "" {
		if date, err := p.parseDate(row[col]); err != nil {
			errors = append(errors, models.ValidationError{
				Field:   "last_resolve_date",
				Value:   row[col],
				Message: fmt.Sprintf("invalid date format: %v", err),
				Row:     rowNumber,
			})
		} else {
			incident.LastResolveDate = &date
		}
	}

	// Parse optional fields
	if col, exists := columnMap["description"]; exists && col < len(row) {
		incident.Description = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["category"]; exists && col < len(row) {
		incident.Category = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["subcategory"]; exists && col < len(row) {
		incident.Subcategory = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["impact"]; exists && col < len(row) {
		incident.Impact = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["urgency"]; exists && col < len(row) {
		incident.Urgency = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["status"]; exists && col < len(row) {
		incident.Status = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["customer_affected"]; exists && col < len(row) {
		incident.CustomerAffected = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["business_service"]; exists && col < len(row) {
		incident.BusinessService = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["root_cause"]; exists && col < len(row) {
		incident.RootCause = strings.TrimSpace(row[col])
	}

	if col, exists := columnMap["resolution_notes"]; exists && col < len(row) {
		incident.ResolutionNotes = strings.TrimSpace(row[col])
	}

	// Calculate resolution time if both dates are available
	incident.CalculateResolutionTime()

	// Set defaults
	incident.SetDefaults()

	// Validate the incident
	if validationErr := incident.ValidateForRow(rowNumber); validationErr != nil {
		if validationErrors, ok := validationErr.(models.ValidationErrors); ok {
			errors = append(errors, validationErrors...)
		} else {
			errors = append(errors, models.ValidationError{
				Field:   "general",
				Value:   "",
				Message: validationErr.Error(),
				Row:     rowNumber,
			})
		}
	}

	return incident, errors
}

// parseDate parses various date formats commonly found in Excel files
func (p *ExcelParser) parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Common date formats
	formats := []string{
		"2006-01-02",           // ISO format
		"01/02/2006",           // US format
		"02/01/2006",           // European format
		"2006/01/02",           // Alternative ISO
		"01-02-2006",           // US with dashes
		"02-01-2006",           // European with dashes
		"2006-01-02 15:04:05",  // ISO with time
		"01/02/2006 15:04:05",  // US with time
		"02/01/2006 15:04:05",  // European with time
	}

	// Try parsing as Excel serial number first
	if serialNum, err := strconv.ParseFloat(dateStr, 64); err == nil {
		// Excel serial date (days since 1900-01-01, with some quirks)
		if serialNum > 0 {
			// Excel incorrectly treats 1900 as a leap year, so we need to adjust
			baseDate := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
			return baseDate.AddDate(0, 0, int(serialNum)), nil
		}
	}

	// Try each format
	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// getDefaultColumnMappings returns the default column name mappings
func getDefaultColumnMappings() map[string]string {
	return map[string]string{
		// Required fields
		"incident_id":        "Incident ID",
		"report_date":        "Report Date",
		"brief_description":  "Brief Description",
		"application_name":   "Application Name",
		"resolution_group":   "Resolution Group",
		"resolved_person":    "Resolved Person",
		"priority":           "Priority",

		// Optional fields
		"resolve_date":       "Resolve Date",
		"last_resolve_date":  "Last Resolve Date",
		"description":        "Description",
		"category":           "Category",
		"subcategory":        "Subcategory",
		"impact":             "Impact",
		"urgency":            "Urgency",
		"status":             "Status",
		"customer_affected":  "Customer Affected",
		"business_service":   "Business Service",
		"root_cause":         "Root Cause",
		"resolution_notes":   "Resolution Notes",
	}
}