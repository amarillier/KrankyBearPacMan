package main

import (
	"strconv"

	"fyne.io/fyne/v2"
)

const maxHighScores = 5

// HighScoreManager manages high scores using Fyne preferences
type HighScoreManager struct {
	prefs fyne.Preferences
}

// NewHighScoreManager creates a new high score manager
func NewHighScoreManager(app fyne.App) *HighScoreManager {
	return &HighScoreManager{
		prefs: app.Preferences(),
	}
}

// GetHighScores retrieves the high scores from preferences
func (h *HighScoreManager) GetHighScores() []int {
	scores := make([]int, 0, maxHighScores)

	// Read scores in order (they should already be sorted descending)
	for i := 0; i < maxHighScores; i++ {
		key := "highscore_" + strconv.Itoa(i)
		score := h.prefs.IntWithFallback(key, 0)
		if score > 0 {
			scores = append(scores, score)
		}
	}

	// Don't sort here - scores are already saved in descending order
	// Sorting here could cause issues if there are any inconsistencies
	return scores
}

// AddScore adds a score if it's high enough, maintaining top 5
// Returns true if this score is a new high score (in top 5)
func (h *HighScoreManager) AddScore(score int) bool {
	if score <= 0 {
		return false
	}

	// Get current high scores (should be sorted descending)
	scores := h.GetHighScores()

	// Check if this score already exists (prevent duplicates)
	for _, s := range scores {
		if s == score {
			return false // Score already exists, don't add again
		}
	}

	// Check if score qualifies for top 5
	isHighScore := len(scores) < maxHighScores || (len(scores) > 0 && score > scores[len(scores)-1])

	if isHighScore {
		// Find insertion point and insert the new score
		inserted := false
		for i := 0; i < len(scores); i++ {
			if score > scores[i] {
				// Insert before this score
				// Create new slice with the score inserted
				newScores := make([]int, 0, maxHighScores+1)
				newScores = append(newScores, scores[:i]...)
				newScores = append(newScores, score)
				newScores = append(newScores, scores[i:]...)
				scores = newScores
				inserted = true
				break
			}
		}
		if !inserted {
			// Append to end if not inserted (score is lowest but still qualifies)
			scores = append(scores, score)
		}

		// Keep only top maxHighScores
		if len(scores) > maxHighScores {
			scores = scores[:maxHighScores]
		}

		// Save back to preferences - always save exactly maxHighScores slots
		// Clear all slots first to avoid duplicates
		for i := 0; i < maxHighScores; i++ {
			key := "highscore_" + strconv.Itoa(i)
			h.prefs.SetInt(key, 0)
		}
		// Now save the scores in order (they should be sorted descending)
		for i := 0; i < len(scores) && i < maxHighScores; i++ {
			key := "highscore_" + strconv.Itoa(i)
			h.prefs.SetInt(key, scores[i])
		}

		return true
	}

	return false
}

// ResetHighScores clears all high scores
func (h *HighScoreManager) ResetHighScores() {
	for i := 0; i < maxHighScores; i++ {
		key := "highscore_" + strconv.Itoa(i)
		h.prefs.SetInt(key, 0)
	}
}
