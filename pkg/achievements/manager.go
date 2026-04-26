package achievements

import (
	"sync"
	"time"
)

// AchievementManager tracks player achievement progress
type AchievementManager struct {
	progress   map[string]*AchievementProgress
	mu         sync.RWMutex
	onUnlocked func(progress *AchievementProgress)
	onProgress func(progress *AchievementProgress)
	stats      *AchievementStats
}

// AchievementStats tracks aggregated stats
type AchievementStats struct {
	TotalAchievements    int
	UnlockedCount        int
	TotalXPEarned        int
	AchievementsByTier   map[AchievementTier]int
	CompletionByCategory map[AchievementCategory]float64
}

// NewAchievementManager creates a new achievement manager
func NewAchievementManager() *AchievementManager {
	am := &AchievementManager{
		progress: make(map[string]*AchievementProgress),
		stats: &AchievementStats{
			AchievementsByTier:   make(map[AchievementTier]int),
			CompletionByCategory: make(map[AchievementCategory]float64),
		},
	}

	// Initialize all achievements
	for id, def := range AchievementRegistry {
		am.progress[id] = &AchievementProgress{
			Definition: def,
			Unlocked:   false,
			Progress:   0,
		}
	}

	am.updateStats()
	return am
}

// SetCallbacks sets event callbacks
func (am *AchievementManager) SetCallbacks(
	onUnlocked func(progress *AchievementProgress),
	onProgress func(progress *AchievementProgress),
) {
	am.onUnlocked = onUnlocked
	am.onProgress = onProgress
}

// UnlockAchievement unlocks an achievement immediately
func (am *AchievementManager) UnlockAchievement(id string) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	progress, exists := am.progress[id]
	if !exists || progress.Unlocked {
		return false
	}

	progress.Unlocked = true
	now := time.Now()
	progress.UnlockedAt = &now

	am.updateStats()

	if am.onUnlocked != nil {
		am.onUnlocked(progress)
	}

	// Check for completionist achievement
	am.checkCompletionist()

	return true
}

// UpdateProgress updates progress for an achievement
func (am *AchievementManager) UpdateProgress(id string, amount int) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	progress, exists := am.progress[id]
	if !exists {
		return false
	}

	// Don't update already unlocked achievements
	if progress.Unlocked {
		return false
	}

	oldProgress := progress.Progress
	progress.Progress += amount

	// Check if complete
	if progress.Definition.IsProgressBased() && progress.Progress >= progress.Definition.MaxProgress {
		progress.Progress = progress.Definition.MaxProgress
		progress.Unlocked = true
		now := time.Now()
		progress.UnlockedAt = &now

		if am.onUnlocked != nil {
			am.onUnlocked(progress)
		}

		am.updateStats()
		am.checkCompletionist()
	} else if am.onProgress != nil && oldProgress != progress.Progress {
		am.onProgress(progress)
	}

	return true
}

// IncrementProgress increments progress by 1
func (am *AchievementManager) IncrementProgress(id string) bool {
	return am.UpdateProgress(id, 1)
}

// GetProgress returns the progress for an achievement
func (am *AchievementManager) GetProgress(id string) *AchievementProgress {
	am.mu.RLock()
	defer am.mu.RUnlock()

	return am.progress[id]
}

// IsUnlocked checks if an achievement is unlocked
func (am *AchievementManager) IsUnlocked(id string) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	progress, exists := am.progress[id]
	if !exists {
		return false
	}
	return progress.Unlocked
}

// GetUnlockedAchievements returns all unlocked achievements
func (am *AchievementManager) GetUnlockedAchievements() []*AchievementProgress {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var unlocked []*AchievementProgress
	for _, progress := range am.progress {
		if progress.Unlocked {
			unlocked = append(unlocked, progress)
		}
	}
	return unlocked
}

// GetProgressByCategory returns achievements in a category
func (am *AchievementManager) GetProgressByCategory(category AchievementCategory) []*AchievementProgress {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var result []*AchievementProgress
	for _, progress := range am.progress {
		if progress.Definition.Category == category {
			result = append(result, progress)
		}
	}
	return result
}

// GetStats returns achievement statistics
func (am *AchievementManager) GetStats() *AchievementStats {
	am.mu.RLock()
	defer am.mu.RUnlock()

	return am.stats
}

// GetCompletionPercentage returns overall completion (0-100)
func (am *AchievementManager) GetCompletionPercentage() float64 {
	am.mu.RLock()
	defer am.mu.RUnlock()

	total := len(am.progress)
	if total == 0 {
		return 0
	}

	unlocked := 0
	for _, progress := range am.progress {
		if progress.Unlocked {
			unlocked++
		}
	}

	return float64(unlocked) / float64(total) * 100
}

// GetRecentUnlocks returns recently unlocked achievements (last N)
func (am *AchievementManager) GetRecentUnlocks(count int) []*AchievementProgress {
	am.mu.RLock()
	defer am.mu.Unlock()

	var unlocked []*AchievementProgress
	for _, progress := range am.progress {
		if progress.Unlocked && progress.UnlockedAt != nil {
			unlocked = append(unlocked, progress)
		}
	}

	// Sort by unlock time (most recent first)
	for i := 0; i < len(unlocked)-1; i++ {
		for j := i + 1; j < len(unlocked); j++ {
			if unlocked[j].UnlockedAt.After(*unlocked[i].UnlockedAt) {
				unlocked[i], unlocked[j] = unlocked[j], unlocked[i]
			}
		}
	}

	if count > 0 && count < len(unlocked) {
		return unlocked[:count]
	}
	return unlocked
}

// Reset resets all achievements
func (am *AchievementManager) Reset() {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, progress := range am.progress {
		progress.Unlocked = false
		progress.Progress = 0
		progress.UnlockedAt = nil
	}

	am.updateStats()
}

// SaveData represents saveable achievement data
type SaveData struct {
	Progress map[string]struct {
		Unlocked   bool       `json:"unlocked"`
		Progress   int        `json:"progress"`
		UnlockedAt *time.Time `json:"unlocked_at,omitempty"`
	} `json:"progress"`
}

// ExportSaveData exports achievement data for saving
func (am *AchievementManager) ExportSaveData() *SaveData {
	am.mu.RLock()
	defer am.mu.RUnlock()

	data := &SaveData{
		Progress: make(map[string]struct {
			Unlocked   bool       `json:"unlocked"`
			Progress   int        `json:"progress"`
			UnlockedAt *time.Time `json:"unlocked_at,omitempty"`
		}),
	}

	for id, progress := range am.progress {
		data.Progress[id] = struct {
			Unlocked   bool       `json:"unlocked"`
			Progress   int        `json:"progress"`
			UnlockedAt *time.Time `json:"unlocked_at,omitempty"`
		}{
			Unlocked:   progress.Unlocked,
			Progress:   progress.Progress,
			UnlockedAt: progress.UnlockedAt,
		}
	}

	return data
}

// ImportSaveData imports achievement data from save
func (am *AchievementManager) ImportSaveData(data *SaveData) {
	am.mu.Lock()
	defer am.mu.Unlock()

	for id, saved := range data.Progress {
		if progress, exists := am.progress[id]; exists {
			progress.Unlocked = saved.Unlocked
			progress.Progress = saved.Progress
			progress.UnlockedAt = saved.UnlockedAt
		}
	}

	am.updateStats()
}

// updateStats updates cached statistics
func (am *AchievementManager) updateStats() {
	am.stats.TotalAchievements = len(am.progress)
	am.stats.UnlockedCount = 0
	am.stats.TotalXPEarned = 0

	// Reset tier counts
	am.stats.AchievementsByTier = make(map[AchievementTier]int)

	// Category completion tracking
	categoryTotal := make(map[AchievementCategory]int)
	categoryUnlocked := make(map[AchievementCategory]int)

	for _, progress := range am.progress {
		categoryTotal[progress.Definition.Category]++

		if progress.Unlocked {
			am.stats.UnlockedCount++
			am.stats.TotalXPEarned += progress.Definition.RewardXP
			am.stats.AchievementsByTier[progress.Definition.Tier]++
			categoryUnlocked[progress.Definition.Category]++
		}
	}

	// Calculate completion percentages by category
	am.stats.CompletionByCategory = make(map[AchievementCategory]float64)
	for cat, total := range categoryTotal {
		if total > 0 {
			am.stats.CompletionByCategory[cat] = float64(categoryUnlocked[cat]) / float64(total) * 100
		}
	}
}

// checkCompletionist checks if completionist achievement should unlock
func (am *AchievementManager) checkCompletionist() {
	completionist, exists := am.progress["completionist"]
	if !exists || completionist.Unlocked {
		return
	}

	// Check if all other achievements are unlocked
	allUnlocked := true
	for id, progress := range am.progress {
		if id != "completionist" && !progress.Unlocked {
			allUnlocked = false
			break
		}
	}

	if allUnlocked {
		completionist.Unlocked = true
		now := time.Now()
		completionist.UnlockedAt = &now

		if am.onUnlocked != nil {
			am.onUnlocked(completionist)
		}
	}
}
