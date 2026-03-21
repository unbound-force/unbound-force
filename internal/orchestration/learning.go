package orchestration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// minPatternOccurrences is the threshold for a pattern to be
// considered significant enough to generate feedback.
const minPatternOccurrences = 3

// AnalyzeWorkflows examines completed workflow records for recurring
// patterns and produces learning feedback recommendations.
//
// Current pattern detectors:
//   - Frequent Divisor findings: same review category > 3 times
//     across workflows → recommend convention pack update
//   - Velocity trends: decreasing review iteration counts →
//     positive feedback on improvement
func AnalyzeWorkflows(records []WorkflowRecord) ([]LearningFeedback, error) {
	if len(records) < minPatternOccurrences {
		return nil, nil
	}

	var feedback []LearningFeedback

	// Pattern 1: Frequent review stage failures/iterations
	reviewPatterns := analyzeReviewPatterns(records)
	feedback = append(feedback, reviewPatterns...)

	// Pattern 2: Velocity improvement trends
	velocityPatterns := analyzeVelocityPatterns(records)
	feedback = append(feedback, velocityPatterns...)

	return feedback, nil
}

// analyzeReviewPatterns looks for recurring review-related issues
// across workflow records. If the review stage consistently has
// errors or high iteration counts, it generates feedback targeting
// cobalt-crush to update coding conventions.
func analyzeReviewPatterns(records []WorkflowRecord) []LearningFeedback {
	// Count review stage errors by error message pattern
	errorCounts := make(map[string]int)
	var workflowIDs []string

	for _, record := range records {
		for _, stage := range record.Stages {
			if stage.StageName == StageReview && stage.Error != "" {
				// Normalize error to a category
				category := normalizeError(stage.Error)
				errorCounts[category]++
				workflowIDs = append(workflowIDs, record.WorkflowID)
			}
		}
	}

	var feedback []LearningFeedback
	for category, count := range errorCounts {
		if count >= minPatternOccurrences {
			feedback = append(feedback, LearningFeedback{
				SourceHero:      "divisor",
				TargetHero:      "cobalt-crush",
				PatternObserved: fmt.Sprintf("review finding %q occurred %d times across %d workflows", category, count, len(records)),
				Recommendation:  fmt.Sprintf("update convention pack to address recurring %q findings proactively", category),
				SupportingData:  map[string]string{"category": category, "count": fmt.Sprintf("%d", count)},
				Status:          "proposed",
				CreatedAt:       time.Now().UTC(),
				WorkflowIDs:     dedup(workflowIDs),
			})
		}
	}

	return feedback
}

// analyzeVelocityPatterns checks if review iteration counts are
// decreasing over time, indicating the team is improving.
func analyzeVelocityPatterns(records []WorkflowRecord) []LearningFeedback {
	if len(records) < minPatternOccurrences {
		return nil
	}

	// Check for decreasing iteration trend in the last N records
	// (records are assumed to be in chronological order)
	improving := true
	for i := 1; i < len(records); i++ {
		prevIterations := countReviewIterations(records[i-1])
		currIterations := countReviewIterations(records[i])
		if currIterations > prevIterations {
			improving = false
			break
		}
	}

	if !improving {
		return nil
	}

	var workflowIDs []string
	for _, r := range records {
		workflowIDs = append(workflowIDs, r.WorkflowID)
	}

	return []LearningFeedback{
		{
			SourceHero:      "mx-f",
			TargetHero:      "cobalt-crush",
			PatternObserved: fmt.Sprintf("review iteration counts decreasing over %d workflows", len(records)),
			Recommendation:  "current coding practices are improving review outcomes — continue current approach",
			SupportingData:  map[string]string{"trend": "improving", "workflows": fmt.Sprintf("%d", len(records))},
			Status:          "proposed",
			CreatedAt:       time.Now().UTC(),
			WorkflowIDs:     workflowIDs,
		},
	}
}

// countReviewIterations counts how many review-related stages
// exist in a workflow record (approximation of iteration count).
func countReviewIterations(record WorkflowRecord) int {
	count := 0
	for _, stage := range record.Stages {
		if stage.StageName == StageReview {
			count++
		}
	}
	return count
}

// normalizeError extracts a category from an error message.
func normalizeError(errMsg string) string {
	// Strip "escalated: " prefix if present
	msg := strings.TrimPrefix(errMsg, "escalated: ")
	// Truncate to first sentence or 50 chars
	if idx := strings.Index(msg, "."); idx > 0 && idx < 50 {
		return msg[:idx]
	}
	if len(msg) > 50 {
		return msg[:50]
	}
	return msg
}

// NextFeedbackID generates the next auto-incrementing LF-NNN ID.
func NextFeedbackID(existing []LearningFeedback) string {
	maxNum := 0
	for _, fb := range existing {
		var num int
		if _, err := fmt.Sscanf(fb.ID, "LF-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}
	return fmt.Sprintf("LF-%03d", maxNum+1)
}

// SaveFeedback writes learning feedback to the learning directory.
// Each feedback item is saved as a separate JSON file.
func SaveFeedback(dir string, feedback []LearningFeedback) error {
	learningDir := filepath.Join(dir, "learning")
	if err := os.MkdirAll(learningDir, 0755); err != nil {
		return fmt.Errorf("create learning directory: %w", err)
	}

	for _, fb := range feedback {
		data, err := json.MarshalIndent(fb, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal feedback %q: %w", fb.ID, err)
		}

		path := filepath.Join(learningDir, fb.ID+".json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("write feedback %q: %w", path, err)
		}
	}

	return nil
}

// LoadFeedback reads all learning feedback from the learning directory.
func LoadFeedback(dir string) ([]LearningFeedback, error) {
	learningDir := filepath.Join(dir, "learning")
	entries, err := os.ReadDir(learningDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read learning directory: %w", err)
	}

	var feedback []LearningFeedback
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(learningDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var fb LearningFeedback
		if err := json.Unmarshal(data, &fb); err != nil {
			continue
		}
		feedback = append(feedback, fb)
	}

	// Sort by ID for deterministic ordering
	sort.Slice(feedback, func(i, j int) bool {
		return feedback[i].ID < feedback[j].ID
	})

	return feedback, nil
}

// dedup removes duplicate strings from a slice.
func dedup(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
