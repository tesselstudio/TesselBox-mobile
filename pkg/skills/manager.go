package skills

import (
	"sync"
	"time"
)

// SkillManager handles player skill progression
type SkillManager struct {
	mu        sync.RWMutex
	skills    map[string]*SkillInstance
	xp        map[SkillTree]int
	totalXP   int
	level     map[SkillTree]int
	onUnlock  func(skill *SkillInstance)
	onLevelUp func(tree SkillTree, level int)
}

// NewSkillManager creates a new skill manager
func NewSkillManager() *SkillManager {
	sm := &SkillManager{
		skills: make(map[string]*SkillInstance),
		xp:     make(map[SkillTree]int),
		level:  make(map[SkillTree]int),
	}

	// Initialize all skills
	for id, def := range SkillRegistry {
		sm.skills[id] = &SkillInstance{
			Definition: def,
			Unlocked:   false,
			Level:      0,
		}
	}

	return sm
}

// SetCallbacks sets event callbacks
func (sm *SkillManager) SetCallbacks(
	onUnlock func(skill *SkillInstance),
	onLevelUp func(tree SkillTree, level int),
) {
	sm.onUnlock = onUnlock
	sm.onLevelUp = onLevelUp
}

// AddXP adds XP to a skill tree
func (sm *SkillManager) AddXP(tree SkillTree, amount int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.xp[tree] += amount
	sm.totalXP += amount

	// Check for level up
	newLevel := sm.calculateLevel(sm.xp[tree])
	if newLevel > sm.level[tree] {
		oldLevel := sm.level[tree]
		sm.level[tree] = newLevel

		if sm.onLevelUp != nil {
			for l := oldLevel + 1; l <= newLevel; l++ {
				sm.onLevelUp(tree, l)
			}
		}
	}
}

// UnlockSkill unlocks a skill
func (sm *SkillManager) UnlockSkill(skillID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	skill, exists := sm.skills[skillID]
	if !exists || skill.Unlocked {
		return false
	}

	// Check prerequisites
	if !skill.CanUnlock(sm.skills) {
		return false
	}

	// Check XP cost
	tree := skill.Definition.Tree
	if sm.xp[tree] < skill.Definition.Cost {
		return false
	}

	// Deduct XP and unlock
	sm.xp[tree] -= skill.Definition.Cost
	skill.Unlocked = true
	skill.Level = 1
	now := time.Now()
	skill.UnlockedAt = &now

	if sm.onUnlock != nil {
		sm.onUnlock(skill)
	}

	return true
}

// GetSkill returns a skill instance
func (sm *SkillManager) GetSkill(skillID string) *SkillInstance {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.skills[skillID]
}

// GetTreeXP returns XP in a tree
func (sm *SkillManager) GetTreeXP(tree SkillTree) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.xp[tree]
}

// GetTreeLevel returns level in a tree
func (sm *SkillManager) GetTreeLevel(tree SkillTree) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.level[tree]
}

// GetUnlockedSkills returns unlocked skills
func (sm *SkillManager) GetUnlockedSkills() []*SkillInstance {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var unlocked []*SkillInstance
	for _, skill := range sm.skills {
		if skill.Unlocked {
			unlocked = append(unlocked, skill)
		}
	}
	return unlocked
}

// GetTreeSkills returns skills in a tree
func (sm *SkillManager) GetTreeSkills(tree SkillTree) []*SkillInstance {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var treeSkills []*SkillInstance
	for _, skill := range sm.skills {
		if skill.Definition.Tree == tree {
			treeSkills = append(treeSkills, skill)
		}
	}
	return treeSkills
}

// GetEffectModifier returns total modifier for an effect type
func (sm *SkillManager) GetEffectModifier(effectType, target string) float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	total := 0.0
	for _, skill := range sm.skills {
		if !skill.Unlocked {
			continue
		}
		for _, effect := range skill.Definition.Effects {
			if effect.Type == effectType && effect.Target == target {
				total += effect.Modifier * float64(skill.Level)
			}
		}
	}
	return total
}

// GetAvailableSkills returns skills that can be unlocked
func (sm *SkillManager) GetAvailableSkills(tree SkillTree) []*SkillInstance {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var available []*SkillInstance
	for _, skill := range sm.skills {
		if skill.Definition.Tree == tree && !skill.Unlocked && skill.CanUnlock(sm.skills) {
			if sm.xp[tree] >= skill.Definition.Cost {
				available = append(available, skill)
			}
		}
	}
	return available
}

// calculateLevel calculates level from XP
func (sm *SkillManager) calculateLevel(xp int) int {
	level := 1
	xpNeeded := 100
	remainingXP := xp

	for remainingXP >= xpNeeded {
		remainingXP -= xpNeeded
		level++
		xpNeeded = int(float64(xpNeeded) * 1.5)
		if level >= 50 {
			break
		}
	}

	return level
}

// SaveData for persistence
type SaveData struct {
	XP     map[SkillTree]int `json:"xp"`
	Skills map[string]int    `json:"skills"` // Level of each skill
}

// Export exports save data
func (sm *SkillManager) Export() *SaveData {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	data := &SaveData{
		XP:     make(map[SkillTree]int),
		Skills: make(map[string]int),
	}

	for tree, xp := range sm.xp {
		data.XP[tree] = xp
	}
	for id, skill := range sm.skills {
		if skill.Unlocked {
			data.Skills[id] = skill.Level
		}
	}

	return data
}

// Import imports save data
func (sm *SkillManager) Import(data *SaveData) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for tree, xp := range data.XP {
		sm.xp[tree] = xp
		sm.level[tree] = sm.calculateLevel(xp)
	}

	for id, level := range data.Skills {
		if skill, exists := sm.skills[id]; exists {
			skill.Unlocked = true
			skill.Level = level
			now := time.Now()
			skill.UnlockedAt = &now
		}
	}
}

// Reset resets all skills
func (sm *SkillManager) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, skill := range sm.skills {
		skill.Unlocked = false
		skill.Level = 0
		skill.UnlockedAt = nil
	}

	for tree := range sm.xp {
		sm.xp[tree] = 0
		sm.level[tree] = 0
	}
	sm.totalXP = 0
}
