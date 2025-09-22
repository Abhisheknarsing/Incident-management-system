package services

import (
	"testing"

	"incident-management-system/internal/models"
)

func TestSimpleAutomationAnalyzer_AnalyzeAutomation(t *testing.T) {
	analyzer := NewSimpleAutomationAnalyzer()

	tests := []struct {
		name              string
		incident          *models.Incident
		expectedFeasible  bool
		expectedGroup     string
		expectError       bool
	}{
		{
			name: "high automation potential - server restart",
			incident: &models.Incident{
				IncidentID:       "INC001",
				BriefDescription: "Server needs restart",
				Description:      "Application server requires restart to resolve memory issue",
				ApplicationName:  "Web Server",
				ResolutionGroup:  "Infrastructure Team",
				Priority:         "P2",
				ResolutionNotes:  "Automated restart script executed successfully",
				ResolutionTimeHours: func() *int { h := 1; return &h }(),
			},
			expectedFeasible: true,
			expectedGroup:    "Infrastructure",
			expectError:      false,
		},
		{
			name: "low automation potential - complex investigation",
			incident: &models.Incident{
				IncidentID:       "INC002",
				BriefDescription: "Complex application error requiring investigation",
				Description:      "Users reporting intermittent errors that require detailed analysis and troubleshooting",
				ApplicationName:  "Custom Business App",
				ResolutionGroup:  "Application Support",
				Priority:         "P3",
				ResolutionNotes:  "Required manual investigation and custom code changes",
				ResolutionTimeHours: func() *int { h := 48; return &h }(),
			},
			expectedFeasible: false,
			expectedGroup:    "Application Support",
			expectError:      false,
		},
		{
			name: "monitoring automation potential",
			incident: &models.Incident{
				IncidentID:       "INC003",
				BriefDescription: "Monitoring alert configuration",
				Description:      "Setup automated monitoring alerts for database performance metrics",
				ApplicationName:  "Database Monitor",
				ResolutionGroup:  "Monitoring Team",
				Priority:         "P4",
				ResolutionNotes:  "Configured automated threshold alerts",
				ResolutionTimeHours: func() *int { h := 2; return &h }(),
			},
			expectedFeasible: true,
			expectedGroup:    "Monitoring",
			expectError:      false,
		},
		{
			name: "security incident - medium automation",
			incident: &models.Incident{
				IncidentID:       "INC004",
				BriefDescription: "Password reset request",
				Description:      "User account password needs to be reset due to security policy",
				ApplicationName:  "Active Directory",
				ResolutionGroup:  "Security Team",
				Priority:         "P3",
				ResolutionNotes:  "Password reset completed through automated system",
				ResolutionTimeHours: func() *int { h := 0; return &h }(), // Very fast
			},
			expectedFeasible: false, // Security typically has lower threshold
			expectedGroup:    "Security",
			expectError:      false,
		},
		{
			name: "user support - low automation",
			incident: &models.Incident{
				IncidentID:       "INC005",
				BriefDescription: "User training request",
				Description:      "New employee needs training on office applications and custom workflows",
				ApplicationName:  "Office Suite",
				ResolutionGroup:  "Help Desk",
				Priority:         "P4",
				ResolutionNotes:  "Provided personalized training session",
				ResolutionTimeHours: func() *int { h := 4; return &h }(),
			},
			expectedFeasible: false,
			expectedGroup:    "User Support",
			expectError:      false,
		},
		{
			name: "nil incident",
			incident: nil,
			expectError: true,
		},
		{
			name: "empty incident",
			incident: &models.Incident{
				IncidentID: "INC006",
			},
			expectedFeasible: false,
			expectedGroup:    "Application Support", // Default
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeAutomation(tt.incident)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if result.Feasible != tt.expectedFeasible {
				t.Errorf("expected feasible %v, got %v (score: %.3f)", 
					tt.expectedFeasible, result.Feasible, result.Score)
			}
			
			if result.ITProcessGroup != tt.expectedGroup {
				t.Errorf("expected IT process group %s, got %s", 
					tt.expectedGroup, result.ITProcessGroup)
			}
			
			// Validate score range
			if result.Score < 0.0 || result.Score > 1.0 {
				t.Errorf("automation score %.3f is outside valid range [0.0, 1.0]", result.Score)
			}
			
			// Validate confidence range
			if result.Confidence < 0.0 || result.Confidence > 1.0 {
				t.Errorf("confidence %.3f is outside valid range [0.0, 1.0]", result.Confidence)
			}
			
			// Check that reasons are provided
			if len(result.Reasons) == 0 {
				t.Errorf("expected reasons to be provided")
			}
		})
	}
}

func TestSimpleAutomationAnalyzer_AnalyzeBatch(t *testing.T) {
	analyzer := NewSimpleAutomationAnalyzer()

	incidents := []*models.Incident{
		{
			IncidentID:       "INC001",
			BriefDescription: "Server restart required",
			Description:      "Automated server restart needed",
			ApplicationName:  "Web Server",
			Priority:         "P2",
		},
		{
			IncidentID:       "INC002",
			BriefDescription: "Complex troubleshooting needed",
			Description:      "Manual investigation and analysis required",
			ApplicationName:  "Custom App",
			Priority:         "P3",
		},
	}

	results, err := analyzer.AnalyzeBatch(incidents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != len(incidents) {
		t.Fatalf("expected %d results, got %d", len(incidents), len(results))
	}

	for i, result := range results {
		// Validate score range
		if result.Score < 0.0 || result.Score > 1.0 {
			t.Errorf("incident %d: automation score %.3f is outside valid range [0.0, 1.0]", 
				i, result.Score)
		}
		
		// Validate confidence range
		if result.Confidence < 0.0 || result.Confidence > 1.0 {
			t.Errorf("incident %d: confidence %.3f is outside valid range [0.0, 1.0]", 
				i, result.Confidence)
		}
		
		// Check IT process group is assigned
		if result.ITProcessGroup == "" {
			t.Errorf("incident %d: IT process group is empty", i)
		}
		
		// Check reasons are provided
		if len(result.Reasons) == 0 {
			t.Errorf("incident %d: no reasons provided", i)
		}
	}

	// First incident should have higher automation potential
	if results[0].Score <= results[1].Score {
		t.Errorf("expected first incident (restart) to have higher automation score than second (troubleshooting)")
	}
}

func TestSimpleAutomationAnalyzer_CategorizeITProcess(t *testing.T) {
	analyzer := NewSimpleAutomationAnalyzer()

	tests := []struct {
		name     string
		incident *models.Incident
		expected string
	}{
		{
			name: "infrastructure keywords",
			incident: &models.Incident{
				BriefDescription: "Server hardware failure",
				Description:      "Database server experiencing hardware issues",
				ApplicationName:  "MySQL Database",
			},
			expected: "Infrastructure",
		},
		{
			name: "security keywords",
			incident: &models.Incident{
				BriefDescription: "Password reset required",
				Description:      "User authentication failed, security certificate expired",
				ApplicationName:  "Active Directory",
			},
			expected: "Security",
		},
		{
			name: "monitoring keywords",
			incident: &models.Incident{
				BriefDescription: "Alert configuration",
				Description:      "Setup monitoring dashboard for performance metrics",
				ApplicationName:  "Monitoring System",
			},
			expected: "Monitoring",
		},
		{
			name: "user support keywords",
			incident: &models.Incident{
				BriefDescription: "User account issue",
				Description:      "Help user with desktop application and printer setup",
				ApplicationName:  "Office Applications",
			},
			expected: "User Support",
		},
		{
			name: "network keywords",
			incident: &models.Incident{
				BriefDescription: "Network connectivity issue",
				Description:      "VPN connection problems, routing configuration needed",
				ApplicationName:  "Network Infrastructure",
			},
			expected: "Network Operations",
		},
		{
			name: "backup keywords",
			incident: &models.Incident{
				BriefDescription: "Backup failure",
				Description:      "Automated backup job failed, restore required",
				ApplicationName:  "Backup System",
			},
			expected: "Backup & Recovery",
		},
		{
			name: "change management keywords",
			incident: &models.Incident{
				BriefDescription: "Software deployment",
				Description:      "Application update and configuration change needed",
				ApplicationName:  "Deployment Pipeline",
			},
			expected: "Change Management",
		},
		{
			name: "empty incident",
			incident: &models.Incident{},
			expected: "Application Support", // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.categorizeITProcess(tt.incident)
			if result != tt.expected {
				t.Errorf("expected IT process group %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSimpleAutomationAnalyzer_AnalyzeTextContent(t *testing.T) {
	analyzer := NewSimpleAutomationAnalyzer()

	tests := []struct {
		name     string
		incident *models.Incident
		expected string // "positive", "negative", or "neutral"
	}{
		{
			name: "automation keywords",
			incident: &models.Incident{
				Description:     "Server restart and automated script execution",
				ResolutionNotes: "Used standard procedure to restart service",
			},
			expected: "positive",
		},
		{
			name: "manual keywords",
			incident: &models.Incident{
				Description:     "Complex troubleshooting and manual investigation required",
				ResolutionNotes: "Escalated to senior engineer for custom analysis",
			},
			expected: "negative",
		},
		{
			name: "neutral content",
			incident: &models.Incident{
				Description: "System status update and general information",
			},
			expected: "neutral",
		},
		{
			name: "empty content",
			incident: &models.Incident{},
			expected: "neutral",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := analyzer.analyzeTextContent(tt.incident)
			
			var actual string
			if score > 0.05 {
				actual = "positive"
			} else if score < -0.05 {
				actual = "negative"
			} else {
				actual = "neutral"
			}
			
			if actual != tt.expected {
				t.Errorf("expected %s sentiment, got %s (score: %.3f)", 
					tt.expected, actual, score)
			}
			
			// Validate score range
			if score < -1.0 || score > 1.0 {
				t.Errorf("text score %.3f is outside valid range [-1.0, 1.0]", score)
			}
		})
	}
}

func TestSimpleAutomationAnalyzer_CalculateResolutionTimeFactor(t *testing.T) {
	analyzer := NewSimpleAutomationAnalyzer()

	tests := []struct {
		name     string
		hours    *int
		expected string // "positive", "negative", or "neutral"
	}{
		{
			name:     "very fast resolution",
			hours:    func() *int { h := 0; return &h }(),
			expected: "positive",
		},
		{
			name:     "fast resolution",
			hours:    func() *int { h := 2; return &h }(),
			expected: "positive",
		},
		{
			name:     "medium resolution",
			hours:    func() *int { h := 12; return &h }(),
			expected: "neutral",
		},
		{
			name:     "slow resolution",
			hours:    func() *int { h := 48; return &h }(),
			expected: "negative",
		},
		{
			name:     "very slow resolution",
			hours:    func() *int { h := 120; return &h }(),
			expected: "negative",
		},
		{
			name:     "no resolution time",
			hours:    nil,
			expected: "neutral",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			incident := &models.Incident{
				ResolutionTimeHours: tt.hours,
			}
			
			factor := analyzer.calculateResolutionTimeFactor(incident)
			
			var actual string
			if factor > 0.05 {
				actual = "positive"
			} else if factor < -0.05 {
				actual = "negative"
			} else {
				actual = "neutral"
			}
			
			if actual != tt.expected {
				t.Errorf("expected %s factor, got %s (factor: %.3f)", 
					tt.expected, actual, factor)
			}
		})
	}
}

func TestSimpleAutomationAnalyzer_CalculatePriorityFactor(t *testing.T) {
	analyzer := NewSimpleAutomationAnalyzer()

	tests := []struct {
		priority string
		expected string // "positive", "negative", or "neutral"
	}{
		{"P1", "positive"},
		{"P2", "positive"},
		{"P3", "neutral"},
		{"P4", "negative"},
		{"", "neutral"},
		{"UNKNOWN", "neutral"},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			factor := analyzer.calculatePriorityFactor(tt.priority)
			
			var actual string
			if factor > 0.05 {
				actual = "positive"
			} else if factor < -0.05 {
				actual = "negative"
			} else {
				actual = "neutral"
			}
			
			if actual != tt.expected {
				t.Errorf("priority %s: expected %s factor, got %s (factor: %.3f)", 
					tt.priority, tt.expected, actual, factor)
			}
		})
	}
}

func TestSimpleAutomationAnalyzer_AddCustomKeywords(t *testing.T) {
	analyzer := NewSimpleAutomationAnalyzer()

	// Add custom keywords
	customAutomation := map[string]float64{
		"automate": 0.9,
		"scripted": 0.8,
	}
	customManual := map[string]float64{
		"investigate": -0.8,
		"analyze": -0.7,
	}

	analyzer.AddCustomKeywords(customAutomation, customManual)

	// Test that custom keywords are recognized
	incident := &models.Incident{
		Description: "This process can be automated using scripted solutions",
	}

	score := analyzer.analyzeTextContent(incident)
	if score <= 0 {
		t.Errorf("expected positive score for custom automation keywords, got %.3f", score)
	}

	incident2 := &models.Incident{
		Description: "Need to investigate and analyze this complex issue",
	}

	score2 := analyzer.analyzeTextContent(incident2)
	if score2 >= 0 {
		t.Errorf("expected negative score for custom manual keywords, got %.3f", score2)
	}
}

func TestBatchProcessIncidentsAutomation(t *testing.T) {
	analyzer := NewSimpleAutomationAnalyzer()

	incidents := []models.Incident{
		{
			IncidentID:       "INC001",
			BriefDescription: "Server restart needed",
			Description:      "Automated restart required for web server",
			Priority:         "P2",
		},
		{
			IncidentID:       "INC002",
			BriefDescription: "Complex investigation",
			Description:      "Manual troubleshooting and analysis required",
			Priority:         "P3",
		},
		{
			IncidentID:       "INC003",
			BriefDescription: "Monitoring setup",
			Description:      "Configure automated monitoring alerts",
			Priority:         "P4",
		},
	}

	err := BatchProcessIncidentsAutomation(analyzer, incidents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that all incidents have automation data
	for i, incident := range incidents {
		if incident.AutomationScore == nil {
			t.Errorf("incident %d: automation score is nil", i)
		} else {
			if *incident.AutomationScore < 0.0 || *incident.AutomationScore > 1.0 {
				t.Errorf("incident %d: invalid automation score %.3f", i, *incident.AutomationScore)
			}
		}

		if incident.AutomationFeasible == nil {
			t.Errorf("incident %d: automation feasible is nil", i)
		}

		if incident.ITProcessGroup == "" {
			t.Errorf("incident %d: IT process group is empty", i)
		}
	}

	// Check expected automation feasibility
	// First incident (restart) should have higher automation potential
	if incidents[0].AutomationScore != nil && incidents[1].AutomationScore != nil {
		if *incidents[0].AutomationScore <= *incidents[1].AutomationScore {
			t.Errorf("expected restart incident to have higher automation score than investigation incident")
		}
	}
}

func TestValidateAutomationScore(t *testing.T) {
	tests := []struct {
		score       float64
		expectError bool
	}{
		{0.0, false},
		{0.5, false},
		{1.0, false},
		{-0.1, true},
		{1.1, true},
		{2.0, true},
		{-1.0, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			err := ValidateAutomationScore(tt.score)
			if tt.expectError && err == nil {
				t.Errorf("expected error for score %.3f", tt.score)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for score %.3f: %v", tt.score, err)
			}
		})
	}
}

func TestSimpleAutomationAnalyzer_GetAutomationStats(t *testing.T) {
	analyzer := NewSimpleAutomationAnalyzer()

	stats := analyzer.GetAutomationStats()

	// Check that stats contain expected keys
	expectedKeys := []string{
		"automation_keywords_count",
		"manual_keywords_count", 
		"it_process_groups_count",
		"it_process_groups",
		"analyzer_type",
	}

	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("expected stats to contain key %s", key)
		}
	}

	// Check that counts are reasonable
	if stats["automation_keywords_count"].(int) == 0 {
		t.Errorf("expected automation keywords count to be > 0")
	}

	if stats["manual_keywords_count"].(int) == 0 {
		t.Errorf("expected manual keywords count to be > 0")
	}

	if stats["it_process_groups_count"].(int) == 0 {
		t.Errorf("expected IT process groups count to be > 0")
	}

	// Check analyzer type
	if stats["analyzer_type"].(string) != "simple_rule_based" {
		t.Errorf("expected analyzer type to be 'simple_rule_based'")
	}
}

func BenchmarkSimpleAutomationAnalyzer_AnalyzeAutomation(b *testing.B) {
	analyzer := NewSimpleAutomationAnalyzer()
	incident := &models.Incident{
		IncidentID:       "INC001",
		BriefDescription: "Server restart required for application recovery",
		Description:      "The web server needs to be restarted to resolve memory issues and restore normal operation",
		ApplicationName:  "Web Server",
		ResolutionGroup:  "Infrastructure Team",
		Priority:         "P2",
		ResolutionNotes:  "Executed automated restart script successfully",
		ResolutionTimeHours: func() *int { h := 1; return &h }(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeAutomation(incident)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkSimpleAutomationAnalyzer_AnalyzeBatch(b *testing.B) {
	analyzer := NewSimpleAutomationAnalyzer()
	incidents := []*models.Incident{
		{
			IncidentID:       "INC001",
			BriefDescription: "Server restart required",
			Description:      "Automated server restart needed",
			Priority:         "P2",
		},
		{
			IncidentID:       "INC002",
			BriefDescription: "Complex troubleshooting",
			Description:      "Manual investigation required",
			Priority:         "P3",
		},
		{
			IncidentID:       "INC003",
			BriefDescription: "Monitoring setup",
			Description:      "Configure automated alerts",
			Priority:         "P4",
		},
		{
			IncidentID:       "INC004",
			BriefDescription: "Security update",
			Description:      "Apply security patches",
			Priority:         "P1",
		},
		{
			IncidentID:       "INC005",
			BriefDescription: "User support",
			Description:      "Help user with application",
			Priority:         "P4",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeBatch(incidents)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}