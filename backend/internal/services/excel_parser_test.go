package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestExcelParser_ParseExcelFile(t *testing.T) {
	// Create a temporary test Excel file
	testFile := createTestExcelFile(t)
	defer os.Remove(testFile)

	parser := NewExcelParser()
	result, err := parser.ParseExcelFile(testFile, "test-upload-id")

	if err != nil {
		t.Fatalf("Failed to parse Excel file: %v", err)
	}

	// Verify results
	if result.TotalRows != 2 {
		t.Errorf("Expected 2 total rows, got %d", result.TotalRows)
	}

	if result.ValidRows != 1 {
		t.Errorf("Expected 1 valid row, got %d", result.ValidRows)
	}

	if len(result.Incidents) != 1 {
		t.Errorf("Expected 1 incident, got %d", len(result.Incidents))
	}

	if len(result.Errors) == 0 {
		t.Error("Expected validation errors for invalid row, got none")
	}

	// Verify the valid incident
	if len(result.Incidents) > 0 {
		incident := result.Incidents[0]
		if incident.IncidentID != "INC001" {
			t.Errorf("Expected incident ID 'INC001', got '%s'", incident.IncidentID)
		}
		if incident.Priority != "P1" {
			t.Errorf("Expected priority 'P1', got '%s'", incident.Priority)
		}
		if incident.ApplicationName != "Test App" {
			t.Errorf("Expected application 'Test App', got '%s'", incident.ApplicationName)
		}
	}
}

func TestExcelParser_ParseDate(t *testing.T) {
	parser := NewExcelParser()

	testCases := []struct {
		input    string
		expected bool // whether parsing should succeed
	}{
		{"2023-01-15", true},
		{"01/15/2023", true},
		{"15/01/2023", true},
		{"2023/01/15", true},
		{"01-15-2023", true},
		{"44927", true}, // Excel serial number for 2023-01-15
		{"invalid-date", false},
		{"", false},
	}

	for _, tc := range testCases {
		_, err := parser.parseDate(tc.input)
		success := err == nil

		if success != tc.expected {
			t.Errorf("parseDate(%s): expected success=%v, got success=%v (error: %v)", 
				tc.input, tc.expected, success, err)
		}
	}
}

func TestExcelParser_NormalizeColumnName(t *testing.T) {
	parser := NewExcelParser()

	testCases := []struct {
		input    string
		expected string
	}{
		{"Incident ID", "incidentid"},
		{"incident_id", "incidentid"},
		{"Incident-ID", "incidentid"},
		{"INCIDENT ID", "incidentid"},
		{"  Incident  ID  ", "incidentid"},
	}

	for _, tc := range testCases {
		result := parser.normalizeColumnName(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeColumnName(%s): expected '%s', got '%s'", 
				tc.input, tc.expected, result)
		}
	}
}

// createTestExcelFile creates a temporary Excel file for testing
func createTestExcelFile(t *testing.T) string {
	f := excelize.NewFile()
	defer f.Close()

	// Create header row
	headers := []string{
		"Incident ID", "Report Date", "Brief Description", "Application Name",
		"Resolution Group", "Resolved Person", "Priority", "Resolve Date",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cell, header)
	}

	// Create valid data row
	validRow := []interface{}{
		"INC001", "2023-01-15", "Test incident description", "Test App",
		"IT Support", "John Doe", "P1", "2023-01-16",
	}

	for i, value := range validRow {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue("Sheet1", cell, value)
	}

	// Create invalid data row (missing required fields)
	invalidRow := []interface{}{
		"", "invalid-date", "", "",
		"", "", "P5", "", // P5 is invalid priority
	}

	for i, value := range invalidRow {
		cell, _ := excelize.CoordinatesToCellName(i+1, 3)
		f.SetCellValue("Sheet1", cell, value)
	}

	// Save to temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.xlsx")
	
	if err := f.SaveAs(testFile); err != nil {
		t.Fatalf("Failed to save test Excel file: %v", err)
	}

	return testFile
}