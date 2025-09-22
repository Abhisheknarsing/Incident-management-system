package services

import (
	"fmt"
	"regexp"
	"strings"

	"incident-management-system/internal/models"
)

// SimpleAutomationAnalyzer implements basic automation analysis
type SimpleAutomationAnalyzer struct {
	automationKeywords    map[string]float64
	manualKeywords        map[string]float64
	itProcessGroups       map[string][]string
	automationThresholds  map[string]float64
	resolutionTimeWeights map[string]float64
}

// NewSimpleAutomationAnalyzer creates a new automation analyzer
func NewSimpleAutomationAnalyzer() *SimpleAutomationAnalyzer {
	analyzer := &SimpleAutomationAnalyzer{
		automationKeywords:    make(map[string]float64),
		manualKeywords:        make(map[string]float64),
		itProcessGroups:       make(map[string][]string),
		automationThresholds:  make(map[string]float64),
		resolutionTimeWeights: make(map[string]float64),
	}

	analyzer.initializeKeywords()
	analyzer.initializeITProcessGroups()
	analyzer.initializeThresholds()
	
	return analyzer
}

// initializeKeywords sets up automation and manual keywords with weights
func (a *SimpleAutomationAnalyzer) initializeKeywords() {
	// Keywords that suggest automation potential (positive weights)
	automationKeywords := map[string]float64{
		"restart":         0.8,
		"reboot":          0.8,
		"reset":           0.7,
		"clear":           0.6,
		"flush":           0.6,
		"refresh":         0.5,
		"reload":          0.6,
		"recycle":         0.7,
		"bounce":          0.7,
		"kill":            0.6,
		"stop":            0.5,
		"start":           0.5,
		"enable":          0.5,
		"disable":         0.5,
		"toggle":          0.6,
		"switch":          0.4,
		"patch":           0.3,
		"install":         0.3,
		"uninstall":       0.4,
		"configure":       0.3,
		"script":          0.7,
		"automated":       0.9,
		"automatic":       0.8,
		"batch":           0.6,
		"scheduled":       0.7,
		"routine":         0.6,
		"standard":        0.5,
		"procedure":       0.4,
		"process":         0.3,
		"workflow":        0.5,
		"template":        0.4,
		"policy":          0.3,
		"rule":            0.4,
		"trigger":         0.6,
		"monitor":         0.4,
		"alert":           0.3,
		"notification":    0.3,
		"backup":          0.6,
		"restore":         0.5,
		"sync":            0.5,
		"synchronize":     0.5,
		"deploy":          0.4,
		"deployment":      0.4,
		"provision":       0.5,
		"cleanup":         0.6,
		"maintenance":     0.4,
		"housekeeping":    0.5,
	}

	// Keywords that suggest manual intervention (negative weights)
	manualKeywords := map[string]float64{
		"investigate":     -0.7,
		"analyze":         -0.6,
		"research":        -0.7,
		"troubleshoot":    -0.8,
		"debug":           -0.7,
		"diagnose":        -0.8,
		"examine":         -0.6,
		"review":          -0.5,
		"inspect":         -0.6,
		"check":           -0.4,
		"verify":          -0.4,
		"validate":        -0.4,
		"test":            -0.3,
		"escalate":        -0.9,
		"escalation":      -0.9,
		"contact":         -0.6,
		"call":            -0.7,
		"email":           -0.5,
		"notify":          -0.4,
		"inform":          -0.4,
		"discuss":         -0.6,
		"meeting":         -0.7,
		"conference":      -0.7,
		"coordinate":      -0.6,
		"collaborate":     -0.5,
		"consult":         -0.6,
		"approve":         -0.6,
		"approval":        -0.6,
		"authorize":       -0.6,
		"permission":      -0.5,
		"manual":          -0.8,
		"manually":        -0.8,
		"custom":          -0.5,
		"customize":       -0.6,
		"personalize":     -0.5,
		"tailor":          -0.5,
		"modify":          -0.4,
		"change":          -0.3,
		"alter":           -0.4,
		"adjust":          -0.4,
		"tweak":           -0.5,
		"fine-tune":       -0.6,
		"complex":         -0.6,
		"complicated":     -0.7,
		"difficult":       -0.6,
		"challenging":     -0.6,
		"unique":          -0.5,
		"special":         -0.4,
		"exception":       -0.7,
		"unusual":         -0.6,
		"rare":            -0.6,
		"one-off":         -0.8,
		"ad-hoc":          -0.7,
	}

	a.automationKeywords = automationKeywords
	a.manualKeywords = manualKeywords
}

// initializeITProcessGroups sets up IT process group classifications
func (a *SimpleAutomationAnalyzer) initializeITProcessGroups() {
	itProcessGroups := map[string][]string{
		"Infrastructure": {
			"server", "servers", "infrastructure", "network", "networking", "hardware",
			"storage", "database", "db", "system", "systems", "platform", "cloud",
			"vm", "virtual", "container", "docker", "kubernetes", "k8s",
			"load balancer", "firewall", "router", "switch", "dns", "dhcp",
			"web server", "application server", "mysql", "postgresql", "oracle",
			"restart", "reboot", "memory", "cpu", "disk", "performance",
		},
		"Application Support": {
			"application", "app", "software", "program", "service", "web",
			"website", "portal", "interface", "ui", "frontend", "backend",
			"api", "microservice", "middleware", "integration", "connector",
			"plugin", "module", "component", "library", "framework",
		},
		"Security": {
			"security", "authentication", "authorization", "access", "permission",
			"certificate", "ssl", "tls", "encryption", "decrypt", "password",
			"credential", "token", "key", "vulnerability", "patch", "antivirus",
			"firewall", "intrusion", "malware", "virus", "threat", "breach",
			"reset", "account", "login", "logout", "active directory", "ad",
			"identity", "policy", "compliance", "audit", "ldap",
		},
		"Monitoring": {
			"monitoring", "monitor", "alert", "notification", "alarm", "dashboard",
			"metric", "log", "logging", "audit", "report", "analytics",
			"performance", "capacity", "utilization", "threshold", "baseline",
			"trend", "anomaly", "health", "status", "availability", "uptime",
		},
		"Backup & Recovery": {
			"backup", "restore", "recovery", "disaster", "failover", "replication",
			"snapshot", "archive", "retention", "rpo", "rto", "continuity",
			"sync", "synchronization", "mirror", "clone", "copy", "dump",
		},
		"Change Management": {
			"change", "deployment", "deploy", "release", "rollback", "rollout",
			"update", "upgrade", "patch", "install", "uninstall", "configure",
			"configuration", "setup", "migration", "maintenance", "schedule",
		},
		"User Support": {
			"user", "users", "account", "profile", "login", "logout", "session",
			"desktop", "laptop", "mobile", "device", "printer", "email",
			"office", "productivity", "training", "onboarding", "offboarding",
			"helpdesk", "support", "ticket", "request", "issue", "problem",
		},
		"Network Operations": {
			"network", "connectivity", "connection", "bandwidth", "latency",
			"routing", "switching", "vlan", "subnet", "ip", "tcp", "udp",
			"port", "protocol", "vpn", "wan", "lan", "wifi", "wireless",
		},
	}

	a.itProcessGroups = itProcessGroups
}

// initializeThresholds sets up automation thresholds and weights
func (a *SimpleAutomationAnalyzer) initializeThresholds() {
	// Automation feasibility thresholds by IT process group
	automationThresholds := map[string]float64{
		"Infrastructure":     0.5, // High automation potential
		"Monitoring":         0.4, // Very high automation potential
		"Backup & Recovery":  0.5, // High automation potential
		"Change Management":  0.5, // Medium automation potential
		"Network Operations": 0.5, // Medium-high automation potential
		"Application Support": 0.4, // Lower automation potential
		"Security":           0.6, // Medium automation potential
		"User Support":       0.3, // Lower automation potential (more human interaction)
	}

	// Resolution time weights (shorter times suggest more automation potential)
	resolutionTimeWeights := map[string]float64{
		"very_fast": 0.3,  // < 1 hour - likely automated or simple
		"fast":      0.2,  // 1-4 hours - good automation candidate
		"medium":    0.0,  // 4-24 hours - neutral
		"slow":      -0.1, // 1-3 days - less automation potential
		"very_slow": -0.2, // > 3 days - likely complex, manual work
	}

	a.automationThresholds = automationThresholds
	a.resolutionTimeWeights = resolutionTimeWeights
}

// AnalyzeAutomation analyzes automation potential for an incident
func (a *SimpleAutomationAnalyzer) AnalyzeAutomation(incident *models.Incident) (*AutomationResult, error) {
	if incident == nil {
		return nil, fmt.Errorf("incident cannot be nil")
	}

	// Calculate resolution time if not already calculated
	if incident.ResolutionTimeHours == nil {
		incident.CalculateResolutionTime()
	}

	// Analyze text content for automation keywords
	textScore := a.analyzeTextContent(incident)
	
	// Determine IT process group
	itProcessGroup := a.categorizeITProcess(incident)
	
	// Get base score for IT process group
	baseScore := a.automationThresholds[itProcessGroup]
	
	// Calculate resolution time factor
	resolutionTimeFactor := a.calculateResolutionTimeFactor(incident)
	
	// Calculate priority factor (higher priority might have more automation potential)
	priorityFactor := a.calculatePriorityFactor(incident.Priority)
	
	// Combine all factors
	finalScore := a.combineFactors(baseScore, textScore, resolutionTimeFactor, priorityFactor)
	
	// Determine feasibility based on IT process group threshold
	threshold := a.automationThresholds[itProcessGroup]
	feasible := finalScore >= threshold
	
	// Calculate confidence based on available data
	confidence := a.calculateConfidence(incident, textScore)
	
	// Generate reasons for the score
	reasons := a.generateReasons(textScore, itProcessGroup, resolutionTimeFactor, priorityFactor, finalScore)

	return &AutomationResult{
		Score:          finalScore,
		Feasible:       feasible,
		ITProcessGroup: itProcessGroup,
		Confidence:     confidence,
		Reasons:        reasons,
	}, nil
}

// AnalyzeBatch analyzes automation potential for multiple incidents
func (a *SimpleAutomationAnalyzer) AnalyzeBatch(incidents []*models.Incident) ([]*AutomationResult, error) {
	results := make([]*AutomationResult, len(incidents))
	
	for i, incident := range incidents {
		result, err := a.AnalyzeAutomation(incident)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze automation for incident %d: %w", i, err)
		}
		results[i] = result
	}
	
	return results, nil
}

// analyzeTextContent analyzes the text content for automation keywords
func (a *SimpleAutomationAnalyzer) analyzeTextContent(incident *models.Incident) float64 {
	// Combine relevant text fields
	text := strings.ToLower(strings.Join([]string{
		incident.BriefDescription,
		incident.Description,
		incident.ResolutionNotes,
		incident.RootCause,
	}, " "))

	if strings.TrimSpace(text) == "" {
		return 0.0
	}

	// Tokenize text
	tokens := a.tokenizeText(text)
	if len(tokens) == 0 {
		return 0.0
	}

	var totalScore float64
	var matchedKeywords int

	// Score automation keywords
	for _, token := range tokens {
		if score, exists := a.automationKeywords[token]; exists {
			totalScore += score
			matchedKeywords++
		}
		if score, exists := a.manualKeywords[token]; exists {
			totalScore += score
			matchedKeywords++
		}
	}

	// Normalize by number of tokens and matched keywords
	if matchedKeywords == 0 {
		return 0.0
	}

	// Average score with some normalization
	avgScore := totalScore / float64(matchedKeywords)
	
	// Apply sigmoid-like normalization to keep in reasonable range
	normalizedScore := avgScore * 0.4 // Scale down to prevent extreme values
	
	// Clamp to [-1, 1] range
	if normalizedScore > 1.0 {
		normalizedScore = 1.0
	} else if normalizedScore < -1.0 {
		normalizedScore = -1.0
	}

	return normalizedScore
}

// categorizeITProcess determines the IT process group for an incident
func (a *SimpleAutomationAnalyzer) categorizeITProcess(incident *models.Incident) string {
	// Combine relevant text fields for classification
	text := strings.ToLower(strings.Join([]string{
		incident.BriefDescription,
		incident.Description,
		incident.ApplicationName,
		incident.ResolutionGroup,
		incident.Category,
		incident.Subcategory,
	}, " "))

	if strings.TrimSpace(text) == "" {
		return "Application Support" // Default category
	}

	// Score each IT process group with weighted scoring
	groupScores := make(map[string]float64)
	
	for group, keywords := range a.itProcessGroups {
		score := 0.0
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				// Give higher weight to longer, more specific keywords
				weight := 1.0
				if len(keyword) > 8 {
					weight = 2.0
				} else if len(keyword) > 5 {
					weight = 1.5
				}
				
				// Give extra weight if keyword appears in brief description or application name
				if strings.Contains(strings.ToLower(incident.BriefDescription), keyword) ||
				   strings.Contains(strings.ToLower(incident.ApplicationName), keyword) {
					weight *= 1.5
				}
				
				score += weight
			}
		}
		groupScores[group] = score
	}

	// Find the group with the highest score
	maxScore := 0.0
	bestGroup := "Application Support" // Default
	
	for group, score := range groupScores {
		if score > maxScore {
			maxScore = score
			bestGroup = group
		}
	}

	return bestGroup
}

// calculateResolutionTimeFactor calculates a factor based on resolution time
func (a *SimpleAutomationAnalyzer) calculateResolutionTimeFactor(incident *models.Incident) float64 {
	if incident.ResolutionTimeHours == nil {
		return 0.0 // No resolution time data
	}

	hours := *incident.ResolutionTimeHours
	
	// Categorize resolution time
	var category string
	switch {
	case hours < 1:
		category = "very_fast"
	case hours < 4:
		category = "fast"
	case hours < 24:
		category = "medium"
	case hours < 72:
		category = "slow"
	default:
		category = "very_slow"
	}

	return a.resolutionTimeWeights[category]
}

// calculatePriorityFactor calculates a factor based on incident priority
func (a *SimpleAutomationAnalyzer) calculatePriorityFactor(priority string) float64 {
	switch strings.ToUpper(priority) {
	case "P1":
		return 0.2 // High priority incidents might benefit from automation
	case "P2":
		return 0.1 // Medium-high priority
	case "P3":
		return 0.0 // Medium priority - neutral
	case "P4":
		return -0.1 // Low priority - less automation benefit
	default:
		return 0.0 // Unknown priority
	}
}

// combineFactors combines all factors into a final automation score
func (a *SimpleAutomationAnalyzer) combineFactors(baseScore, textScore, resolutionTimeFactor, priorityFactor float64) float64 {
	// Weighted combination of factors
	weights := map[string]float64{
		"base":            0.4, // IT process group base score
		"text":            0.4, // Text analysis score
		"resolution_time": 0.15, // Resolution time factor
		"priority":        0.05, // Priority factor
	}

	finalScore := baseScore*weights["base"] +
		textScore*weights["text"] +
		resolutionTimeFactor*weights["resolution_time"] +
		priorityFactor*weights["priority"]

	// Add bonus for positive text score
	if textScore > 0.05 {
		finalScore += 0.15 // Bonus for automation-friendly keywords
	}

	// Ensure score is in [0, 1] range
	if finalScore < 0.0 {
		finalScore = 0.0
	} else if finalScore > 1.0 {
		finalScore = 1.0
	}

	return finalScore
}

// calculateConfidence calculates confidence in the automation analysis
func (a *SimpleAutomationAnalyzer) calculateConfidence(incident *models.Incident, textScore float64) float64 {
	confidence := 0.5 // Base confidence

	// Increase confidence based on available data
	if incident.ResolutionTimeHours != nil {
		confidence += 0.2
	}
	if strings.TrimSpace(incident.Description) != "" {
		confidence += 0.1
	}
	if strings.TrimSpace(incident.ResolutionNotes) != "" {
		confidence += 0.1
	}
	if strings.TrimSpace(incident.RootCause) != "" {
		confidence += 0.1
	}
	if textScore != 0.0 {
		confidence += 0.1 // Text analysis found relevant keywords
	}

	// Clamp to [0, 1] range
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// generateReasons generates human-readable reasons for the automation score
func (a *SimpleAutomationAnalyzer) generateReasons(textScore float64, itProcessGroup string, resolutionTimeFactor, priorityFactor, finalScore float64) []string {
	var reasons []string

	// IT Process Group reason
	reasons = append(reasons, fmt.Sprintf("Categorized as %s (base automation potential: %.1f)", 
		itProcessGroup, a.automationThresholds[itProcessGroup]))

	// Text analysis reason
	if textScore > 0.1 {
		reasons = append(reasons, "Description contains automation-friendly keywords")
	} else if textScore < -0.1 {
		reasons = append(reasons, "Description suggests manual intervention required")
	} else {
		reasons = append(reasons, "Description is neutral regarding automation potential")
	}

	// Resolution time reason
	if resolutionTimeFactor > 0.1 {
		reasons = append(reasons, "Fast resolution time suggests automation potential")
	} else if resolutionTimeFactor < -0.1 {
		reasons = append(reasons, "Slow resolution time suggests complex manual work")
	}

	// Priority reason
	if priorityFactor > 0.05 {
		reasons = append(reasons, "High priority incidents benefit from automation")
	}

	// Final assessment
	if finalScore >= 0.7 {
		reasons = append(reasons, "High automation potential - strongly recommended")
	} else if finalScore >= 0.5 {
		reasons = append(reasons, "Moderate automation potential - worth investigating")
	} else if finalScore >= 0.3 {
		reasons = append(reasons, "Low automation potential - manual process preferred")
	} else {
		reasons = append(reasons, "Very low automation potential - requires human expertise")
	}

	return reasons
}

// tokenizeText tokenizes text for keyword analysis
func (a *SimpleAutomationAnalyzer) tokenizeText(text string) []string {
	// Remove punctuation and split into words
	reg := regexp.MustCompile(`[^\p{L}\p{N}\s-]`)
	cleanText := reg.ReplaceAllString(text, " ")
	
	// Split into words and filter
	words := strings.Fields(cleanText)
	var tokens []string
	
	for _, word := range words {
		word = strings.TrimSpace(strings.ToLower(word))
		if len(word) >= 2 { // Keep words with 2+ characters
			tokens = append(tokens, word)
		}
	}
	
	return tokens
}

// GetAutomationStats returns statistics about the automation analyzer
func (a *SimpleAutomationAnalyzer) GetAutomationStats() map[string]interface{} {
	return map[string]interface{}{
		"automation_keywords_count": len(a.automationKeywords),
		"manual_keywords_count":     len(a.manualKeywords),
		"it_process_groups_count":   len(a.itProcessGroups),
		"it_process_groups":         a.getITProcessGroupNames(),
		"analyzer_type":             "simple_rule_based",
	}
}

// getITProcessGroupNames returns the names of all IT process groups
func (a *SimpleAutomationAnalyzer) getITProcessGroupNames() []string {
	var names []string
	for name := range a.itProcessGroups {
		names = append(names, name)
	}
	return names
}

// AddCustomKeywords allows adding custom automation keywords
func (a *SimpleAutomationAnalyzer) AddCustomKeywords(automation, manual map[string]float64) {
	for word, score := range automation {
		a.automationKeywords[strings.ToLower(word)] = score
	}
	
	for word, score := range manual {
		a.manualKeywords[strings.ToLower(word)] = score
	}
}

// ValidateAutomationScore ensures automation scores are within valid range
func ValidateAutomationScore(score float64) error {
	if score < 0.0 || score > 1.0 {
		return fmt.Errorf("automation score %.3f is outside valid range [0.0, 1.0]", score)
	}
	return nil
}

// BatchProcessIncidentsAutomation processes automation analysis for a batch of incidents
func BatchProcessIncidentsAutomation(analyzer AutomationAnalyzer, incidents []models.Incident) error {
	for i := range incidents {
		// Analyze automation potential
		result, err := analyzer.AnalyzeAutomation(&incidents[i])
		if err != nil {
			return fmt.Errorf("failed to analyze automation for incident %s: %w", incidents[i].IncidentID, err)
		}
		
		// Validate results
		if err := ValidateAutomationScore(result.Score); err != nil {
			return fmt.Errorf("invalid automation score for incident %s: %w", incidents[i].IncidentID, err)
		}
		
		// Update incident
		incidents[i].AutomationScore = &result.Score
		incidents[i].AutomationFeasible = &result.Feasible
		incidents[i].ITProcessGroup = result.ITProcessGroup
	}
	
	return nil
}