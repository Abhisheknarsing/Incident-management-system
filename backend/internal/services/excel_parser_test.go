package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExcelParser_NewExcelParser(t *testing.T) {
	// Test creating parser with nil config (should use defaults)
	parser := NewExcelParser(nil)
	assert.NotNil(t, parser)
	assert.Equal(t, DefaultExcelParserConfig().MaxWorkers, parser.maxWorkers)
	assert.Equal(t, DefaultExcelParserConfig().BatchSize, parser.batchSize)

	// Test creating parser with custom config
	config := &ExcelParserConfig{
		MaxWorkers: 4,
		BatchSize:  50,
	}
	parser = NewExcelParser(config)
	assert.NotNil(t, parser)
	assert.Equal(t, 4, parser.maxWorkers)
	assert.Equal(t, 50, parser.batchSize)

	// Test creating parser with invalid config (should use defaults)
	invalidConfig := &ExcelParserConfig{
		MaxWorkers: -1,
		BatchSize:  -1,
	}
	parser = NewExcelParser(invalidConfig)
	assert.NotNil(t, parser)
	assert.Equal(t, 1, parser.maxWorkers) // Should be corrected to 1
	assert.Equal(t, 10, parser.batchSize) // Should be corrected to 10
}

func TestExcelParser_ParseFile(t *testing.T) {
	parser := NewExcelParser(nil)

	// Test parsing non-existent file
	_, err := parser.ParseFile(context.Background(), "non_existent_file.xlsx")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open Excel file")
}

func TestExcelParser_ParseHeader(t *testing.T) {
	parser := NewExcelParser(nil)

	// Test various header formats
	testCases := []struct {
		name     string
		header   []string
		expected map[string]int
	}{
		{
			name:   "Standard headers",
			header: []string{"incident_id", "application_name", "report_date"},
			expected: map[string]int{
				"incident_id":      0,
				"application_name": 1,
				"report_date":      2,
			},
		},
		{
			name:   "Alternative headers",
			header: []string{"ticket_id", "app", "date"},
			expected: map[string]int{
				"incident_id":      0,
				"application_name": 1,
				"report_date":      2,
			},
		},
		{
			name:   "Mixed case headers",
			header: []string{"Incident ID", "Application Name", "Report Date"},
			expected: map[string]int{
				"incident_id":      0,
				"application_name": 1,
				"report_date":      2,
			},
		},
		{
			name:   "Headers with spaces and underscores",
			header: []string{"incident id", "application_name", "report-date"},
			expected: map[string]int{
				"incident_id":      0,
				"application_name": 1,
				"report_date":      2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			indices := parser.parseHeader(tc.header)
			for expectedKey, expectedIndex := range tc.expected {
				actualIndex, exists := indices[expectedKey]
				assert.True(t, exists, "Expected key %s not found", expectedKey)
				assert.Equal(t, expectedIndex, actualIndex, "Index mismatch for key %s", expectedKey)
			}
		})
	}
}

func TestExcelParser_NormalizeColumnName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"incident_id", "incidentid"},
		{"Incident ID", "incidentid"},
		{"incident-id", "incidentid"},
		{"incident_id_test", "incidentidtest"},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeColumnName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExcelParser_ParseDate(t *testing.T) {
	testCases := []struct {
		name        string
		dateStr     string
		expected    time.Time
		shouldError bool
	}{
		{
			name:        "ISO date format",
			dateStr:     "2024-01-15",
			expected:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "US date format",
			dateStr:     "01/15/2024",
			expected:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "European date format",
			dateStr:     "15/01/2024",
			expected:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "Invalid date format",
			dateStr:     "invalid-date",
			expected:    time.Time{},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseDate(tc.dateStr)

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.Format("2006-01-02"), result.Format("2006-01-02"))
			}
		})
	}
}
