package services

import (
	"fmt"
	"regexp"
	"strings"

	"incident-management-system/internal/models"
)

// SimpleSentimentAnalyzer implements basic sentiment analysis
type SimpleSentimentAnalyzer struct {
	positiveWords map[string]float64
	negativeWords map[string]float64
	intensifiers  map[string]float64
	negators      map[string]bool
}

// NewSimpleSentimentAnalyzer creates a new simple sentiment analyzer
func NewSimpleSentimentAnalyzer() *SimpleSentimentAnalyzer {
	analyzer := &SimpleSentimentAnalyzer{
		positiveWords: make(map[string]float64),
		negativeWords: make(map[string]float64),
		intensifiers:  make(map[string]float64),
		negators:      make(map[string]bool),
	}

	analyzer.initializeWordLists()
	return analyzer
}

// initializeWordLists initializes the sentiment word lists
func (s *SimpleSentimentAnalyzer) initializeWordLists() {
	// Positive words with weights
	positiveWords := map[string]float64{
		"resolved":    0.8,
		"fixed":       0.7,
		"working":     0.6,
		"successful":  0.8,
		"completed":   0.7,
		"good":        0.5,
		"excellent":   0.9,
		"perfect":     0.9,
		"great":       0.7,
		"awesome":     0.8,
		"fantastic":   0.8,
		"wonderful":   0.7,
		"amazing":     0.8,
		"outstanding": 0.9,
		"superb":      0.8,
		"brilliant":   0.8,
		"effective":   0.6,
		"efficient":   0.6,
		"quick":       0.5,
		"fast":        0.5,
		"smooth":      0.6,
		"stable":      0.6,
		"reliable":    0.7,
		"satisfied":   0.6,
		"happy":       0.6,
		"pleased":     0.6,
		"impressed":   0.7,
	}

	// Negative words with weights
	negativeWords := map[string]float64{
		"failed":      -0.8,
		"error":       -0.6,
		"broken":      -0.7,
		"issue":       -0.5,
		"problem":     -0.6,
		"bug":         -0.6,
		"crash":       -0.8,
		"down":        -0.7,
		"offline":     -0.7,
		"unavailable": -0.7,
		"slow":        -0.5,
		"timeout":     -0.6,
		"freeze":      -0.6,
		"hang":        -0.6,
		"stuck":       -0.6,
		"bad":         -0.5,
		"terrible":    -0.9,
		"awful":       -0.8,
		"horrible":    -0.8,
		"worst":       -0.9,
		"hate":        -0.8,
		"frustrated":  -0.7,
		"annoyed":     -0.6,
		"angry":       -0.7,
		"upset":       -0.6,
		"disappointed": -0.7,
		"confused":    -0.5,
		"lost":        -0.5,
		"critical":    -0.8,
		"urgent":      -0.6,
		"emergency":   -0.8,
		"outage":      -0.8,
		"failure":     -0.7,
		"malfunction": -0.7,
		"defect":      -0.6,
		"fault":       -0.6,
	}

	// Intensifiers that modify sentiment strength
	intensifiers := map[string]float64{
		"very":        1.5,
		"extremely":   2.0,
		"really":      1.3,
		"quite":       1.2,
		"totally":     1.8,
		"completely":  1.8,
		"absolutely":  2.0,
		"incredibly":  1.8,
		"amazingly":   1.7,
		"seriously":   1.5,
		"definitely":  1.4,
		"certainly":   1.3,
		"particularly": 1.3,
		"especially":  1.4,
		"highly":      1.4,
		"severely":    1.6,
		"critically":  1.8,
		"urgently":    1.5,
	}

	// Negators that flip sentiment
	negators := map[string]bool{
		"not":     true,
		"no":      true,
		"never":   true,
		"nothing": true,
		"nobody":  true,
		"nowhere": true,
		"neither": true,
		"nor":     true,
		"none":    true,
		"without": true,
		"lack":    true,
		"missing": true,
		"absent":  true,
		"unable":  true,
		"cannot":  true,
		"can't":   true,
		"won't":   true,
		"don't":   true,
		"doesn't": true,
		"didn't":  true,
		"isn't":   true,
		"aren't":  true,
		"wasn't":  true,
		"weren't": true,
		"hasn't":  true,
		"haven't": true,
		"hadn't":  true,
	}

	s.positiveWords = positiveWords
	s.negativeWords = negativeWords
	s.intensifiers = intensifiers
	s.negators = negators
}

// AnalyzeSentiment analyzes the sentiment of a given text
func (s *SimpleSentimentAnalyzer) AnalyzeSentiment(text string) (*SentimentResult, error) {
	if strings.TrimSpace(text) == "" {
		return &SentimentResult{
			Score: 0.0,
			Label: models.SentimentNeutral,
		}, nil
	}

	// Normalize and tokenize text
	tokens := s.tokenize(text)
	if len(tokens) == 0 {
		return &SentimentResult{
			Score: 0.0,
			Label: models.SentimentNeutral,
		}, nil
	}

	// Calculate sentiment score
	score := s.calculateSentimentScore(tokens)

	// Normalize score to [-1, 1] range
	normalizedScore := s.normalizeScore(score, len(tokens))

	// Determine sentiment label
	label := s.scoreToLabel(normalizedScore)

	return &SentimentResult{
		Score: normalizedScore,
		Label: label,
	}, nil
}

// AnalyzeBatch analyzes sentiment for multiple texts
func (s *SimpleSentimentAnalyzer) AnalyzeBatch(texts []string) ([]*SentimentResult, error) {
	results := make([]*SentimentResult, len(texts))
	
	for i, text := range texts {
		result, err := s.AnalyzeSentiment(text)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze sentiment for text %d: %w", i, err)
		}
		results[i] = result
	}
	
	return results, nil
}

// tokenize breaks text into tokens and normalizes them
func (s *SimpleSentimentAnalyzer) tokenize(text string) []string {
	// Convert to lowercase
	text = strings.ToLower(text)
	
	// Remove punctuation except apostrophes (for contractions)
	reg := regexp.MustCompile(`[^\p{L}\p{N}\s']`)
	text = reg.ReplaceAllString(text, " ")
	
	// Split into words
	words := strings.Fields(text)
	
	// Filter out very short words and normalize
	var tokens []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) >= 2 { // Keep words with 2+ characters
			tokens = append(tokens, word)
		}
	}
	
	return tokens
}

// calculateSentimentScore calculates the raw sentiment score
func (s *SimpleSentimentAnalyzer) calculateSentimentScore(tokens []string) float64 {
	var totalScore float64
	var intensifier float64 = 1.0
	var negated bool = false
	
	for i, token := range tokens {
		// Check for intensifiers
		if intensity, isIntensifier := s.intensifiers[token]; isIntensifier {
			intensifier = intensity
			continue
		}
		
		// Check for negators
		if s.negators[token] {
			negated = true
			continue
		}
		
		// Check for sentiment words
		var wordScore float64
		var foundSentiment bool
		
		if score, isPositive := s.positiveWords[token]; isPositive {
			wordScore = score
			foundSentiment = true
		} else if score, isNegative := s.negativeWords[token]; isNegative {
			wordScore = score
			foundSentiment = true
		}
		
		if foundSentiment {
			// Apply intensifier
			wordScore *= intensifier
			
			// Apply negation
			if negated {
				wordScore *= -1
			}
			
			totalScore += wordScore
			
			// Reset modifiers after applying them
			intensifier = 1.0
			negated = false
		} else {
			// Reset modifiers if no sentiment word follows
			if i > 0 && (intensifier != 1.0 || negated) {
				intensifier = 1.0
				negated = false
			}
		}
	}
	
	return totalScore
}

// normalizeScore normalizes the sentiment score to [-1, 1] range
func (s *SimpleSentimentAnalyzer) normalizeScore(score float64, tokenCount int) float64 {
	if tokenCount == 0 {
		return 0.0
	}
	
	// Use a scaling factor to make sentiment more sensitive
	// This amplifies the sentiment while keeping it in bounds
	scalingFactor := 2.0
	normalizedScore := score * scalingFactor
	
	// Clamp to [-1, 1] range
	if normalizedScore > 1.0 {
		normalizedScore = 1.0
	} else if normalizedScore < -1.0 {
		normalizedScore = -1.0
	}
	
	return normalizedScore
}

// scoreToLabel converts a sentiment score to a label
func (s *SimpleSentimentAnalyzer) scoreToLabel(score float64) string {
	if score > 0.05 {
		return models.SentimentPositive
	} else if score < -0.05 {
		return models.SentimentNegative
	} else {
		return models.SentimentNeutral
	}
}

// GetSentimentStats returns statistics about the sentiment analysis
func (s *SimpleSentimentAnalyzer) GetSentimentStats() map[string]interface{} {
	return map[string]interface{}{
		"positive_words_count": len(s.positiveWords),
		"negative_words_count": len(s.negativeWords),
		"intensifiers_count":   len(s.intensifiers),
		"negators_count":       len(s.negators),
		"analyzer_type":        "simple_rule_based",
	}
}

// AddCustomWords allows adding custom sentiment words
func (s *SimpleSentimentAnalyzer) AddCustomWords(positive, negative map[string]float64) {
	for word, score := range positive {
		s.positiveWords[strings.ToLower(word)] = score
	}
	
	for word, score := range negative {
		s.negativeWords[strings.ToLower(word)] = score
	}
}

// ValidateScore ensures sentiment scores are within valid range
func ValidateSentimentScore(score float64) error {
	if score < -1.0 || score > 1.0 {
		return fmt.Errorf("sentiment score %.3f is outside valid range [-1.0, 1.0]", score)
	}
	return nil
}

// ValidateLabel ensures sentiment labels are valid
func ValidateSentimentLabel(label string) error {
	validLabels := []string{models.SentimentPositive, models.SentimentNegative, models.SentimentNeutral}
	for _, valid := range validLabels {
		if label == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid sentiment label '%s', must be one of: %v", label, validLabels)
}

// BatchProcessIncidents processes sentiment analysis for a batch of incidents
func BatchProcessIncidents(analyzer SentimentAnalyzer, incidents []models.Incident) error {
	for i := range incidents {
		// Analyze description field
		result, err := analyzer.AnalyzeSentiment(incidents[i].Description)
		if err != nil {
			return fmt.Errorf("failed to analyze sentiment for incident %s: %w", incidents[i].IncidentID, err)
		}
		
		// Validate results
		if err := ValidateSentimentScore(result.Score); err != nil {
			return fmt.Errorf("invalid sentiment score for incident %s: %w", incidents[i].IncidentID, err)
		}
		
		if err := ValidateSentimentLabel(result.Label); err != nil {
			return fmt.Errorf("invalid sentiment label for incident %s: %w", incidents[i].IncidentID, err)
		}
		
		// Update incident
		incidents[i].SentimentScore = &result.Score
		incidents[i].SentimentLabel = result.Label
	}
	
	return nil
}