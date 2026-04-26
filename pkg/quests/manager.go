package quests

import (
	"sync"
	"time"
)

// QuestManager handles quest state
type QuestManager struct {
	mu            sync.RWMutex
	active        []*QuestInstance
	completed     []*QuestInstance
	available     []*QuestDefinition
	onQuestStart  func(quest *QuestInstance)
	onQuestComplete func(quest *QuestInstance)
	onQuestFail   func(quest *QuestInstance)
}

// NewQuestManager creates a new quest manager
func NewQuestManager() *QuestManager {
	qm := &QuestManager{
		active:    make([]*QuestInstance, 0),
		completed: make([]*QuestInstance, 0),
		available: make([]*QuestDefinition, 0),
	}
	
	// Initialize available quests
	for _, def := range QuestRegistry {
		if len(def.Prereqs) == 0 {
			qm.available = append(qm.available, def)
		}
	}
	
	return qm
}

// SetCallbacks sets event callbacks
func (qm *QuestManager) SetCallbacks(
	onQuestStart func(quest *QuestInstance),
	onQuestComplete func(quest *QuestInstance),
	onQuestFail func(quest *QuestInstance),
) {
	qm.onQuestStart = onQuestStart
	qm.onQuestComplete = onQuestComplete
	qm.onQuestFail = onQuestFail
}

// AcceptQuest accepts an available quest
func (qm *QuestManager) AcceptQuest(questID string) *QuestInstance {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	
	// Find in available
	var def *QuestDefinition
	for i, d := range qm.available {
		if d.ID == questID {
			def = d
			qm.available = append(qm.available[:i], qm.available[i+1:]...)
			break
		}
	}
	
	if def == nil {
		return nil
	}
	
	// Check prerequisites
	for _, prereqID := range def.Prereqs {
		hasPrereq := false
		for _, comp := range qm.completed {
			if comp.Definition.ID == prereqID {
				hasPrereq = true
				break
			}
		}
		if !hasPrereq {
			// Put back in available
			qm.available = append(qm.available, def)
			return nil
		}
	}
	
	// Create instance
	instance := &QuestInstance{
		Definition: def,
		Objectives: make([]QuestObjective, len(def.Objectives)),
		AcceptedAt: time.Now(),
	}
	copy(instance.Objectives, def.Objectives)
	
	qm.active = append(qm.active, instance)
	
	if qm.onQuestStart != nil {
		qm.onQuestStart(instance)
	}
	
	return instance
}

// UpdateObjective updates an objective on all active quests
func (qm *QuestManager) UpdateObjective(objType, target string, amount int) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	
	for _, quest := range qm.active {
		if quest.UpdateObjective(objType, target, amount) {
			if quest.IsComplete() {
				quest.Complete()
				qm.completeQuest(quest)
			}
		}
	}
}

// Update updates all active quests
func (qm *QuestManager) Update(deltaTime float64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	
	for _, quest := range qm.active {
		if quest.Definition.TimeLimit > 0 {
			if quest.CheckTimeLimit() {
				if qm.onQuestFail != nil {
					qm.onQuestFail(quest)
				}
			}
		}
	}
}

// completeQuest moves quest to completed
func (qm *QuestManager) completeQuest(quest *QuestInstance) {
	for i, a := range qm.active {
		if a == quest {
			qm.active = append(qm.active[:i], qm.active[i+1:]...)
			break
		}
	}
	
	qm.completed = append(qm.completed, quest)
	
	// Unlock new quests based on prerequisites
	for _, def := range QuestRegistry {
		if def.HasPrereq(quest.Definition.ID) && !qm.isQuestAvailableOrCompleted(def.ID) {
			qm.available = append(qm.available, def)
		}
	}
	
	if qm.onQuestComplete != nil {
		qm.onQuestComplete(quest)
	}
}

// isQuestAvailableOrCompleted checks quest state
func (qm *QuestManager) isQuestAvailableOrCompleted(id string) bool {
	for _, a := range qm.available {
		if a.ID == id {
			return true
		}
	}
	for _, c := range qm.completed {
		if c.Definition.ID == id {
			return true
		}
	}
	for _, a := range qm.active {
		if a.Definition.ID == id {
			return true
		}
	}
	return false
}

// GetActiveQuests returns active quests
func (qm *QuestManager) GetActiveQuests() []*QuestInstance {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	
	return append([]*QuestInstance{}, qm.active...)
}

// GetAvailableQuests returns available quests
func (qm *QuestManager) GetAvailableQuests() []*QuestDefinition {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	
	return append([]*QuestDefinition{}, qm.available...)
}

// GetCompletedQuests returns completed quests
func (qm *QuestManager) GetCompletedQuests() []*QuestInstance {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	
	return append([]*QuestInstance{}, qm.completed...)
}

// GetQuestProgress returns progress for a quest
func (qm *QuestManager) GetQuestProgress(questID string) *QuestInstance {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	
	for _, q := range qm.active {
		if q.Definition.ID == questID {
			return q
		}
	}
	return nil
}

// GetTotalRewards returns sum of all completed quest rewards
func (qm *QuestManager) GetTotalRewards() (xp, gold int) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	
	for _, q := range qm.completed {
		xp += q.Definition.Reward.XP
		gold += q.Definition.Reward.Gold
	}
	return
}

// AbandonQuest abandons an active quest
func (qm *QuestManager) AbandonQuest(questID string) bool {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	
	for i, q := range qm.active {
		if q.Definition.ID == questID {
			qm.active = append(qm.active[:i], qm.active[i+1:]...)
			// Return to available
			qm.available = append(qm.available, q.Definition)
			return true
		}
	}
	return false
}

// HasPrereq checks if this quest is a prerequisite for any active/available quest
func (qd *QuestDefinition) HasPrereq(questID string) bool {
	for _, prereq := range qd.Prereqs {
		if prereq == questID {
			return true
		}
	}
	return false
}
