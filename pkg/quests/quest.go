package quests

import (
	"time"
)

// QuestType represents quest categories
type QuestType int

const (
	QUEST_GATHER QuestType = iota
	QUEST_CRAFT
	QUEST_EXPLORE
	QUEST_SURVIVE
	QUEST_COMBAT
	QUEST_BUILD
)

func (qt QuestType) String() string {
	switch qt {
	case QUEST_GATHER:
		return "Gather"
	case QUEST_CRAFT:
		return "Craft"
	case QUEST_EXPLORE:
		return "Explore"
	case QUEST_SURVIVE:
		return "Survive"
	case QUEST_COMBAT:
		return "Combat"
	case QUEST_BUILD:
		return "Build"
	default:
		return "Unknown"
	}
}

// QuestObjective represents a quest requirement
type QuestObjective struct {
	Type     string // "gather", "craft", "kill", "explore", "survive"
	Target   string // item ID, enemy type, etc.
	Quantity int
	Current  int
}

// QuestReward represents quest completion rewards
type QuestReward struct {
	XP        int
	Gold      int
	Items     map[string]int // item ID -> quantity
	Reputation int
}

// QuestDefinition defines a quest
type QuestDefinition struct {
	ID          string
	Name        string
	Description string
	Type        QuestType
	Objectives  []QuestObjective
	Reward      QuestReward
	Prereqs     []string // Required quest IDs
	TimeLimit   time.Duration // 0 = no limit
	ChainID     string // Quest chain this belongs to
	ChainOrder  int    // Position in chain
}

// IsComplete checks if all objectives are met
func (qd *QuestDefinition) IsComplete(objectives []QuestObjective) bool {
	if len(objectives) != len(qd.Objectives) {
		return false
	}
	for i, obj := range objectives {
		if obj.Current < qd.Objectives[i].Quantity {
			return false
		}
	}
	return true
}

// QuestInstance represents an active quest
type QuestInstance struct {
	Definition   *QuestDefinition
	Objectives   []QuestObjective
	AcceptedAt   time.Time
	CompletedAt  *time.Time
	ExpiredAt    *time.Time
	Completed    bool
	Failed       bool
}

// UpdateObjective updates an objective
func (qi *QuestInstance) UpdateObjective(objType, target string, amount int) bool {
	if qi.Completed || qi.Failed {
		return false
	}
	
	for i, obj := range qi.Objectives {
		if obj.Type == objType && obj.Target == target {
			qi.Objectives[i].Current += amount
			if qi.Objectives[i].Current > obj.Quantity {
				qi.Objectives[i].Current = obj.Quantity
			}
			return true
		}
	}
	return false
}

// IsComplete checks if quest is complete
func (qi *QuestInstance) IsComplete() bool {
	return qi.Definition.IsComplete(qi.Objectives)
}

// GetProgress returns completion percentage
func (qi *QuestInstance) GetProgress() float64 {
	if len(qi.Objectives) == 0 {
		return 100.0
	}
	
	total := 0
	current := 0
	for i, obj := range qi.Objectives {
		total += qi.Definition.Objectives[i].Quantity
		current += obj.Current
	}
	
	return float64(current) / float64(total) * 100
}

// CheckTimeLimit checks if quest has expired
func (qi *QuestInstance) CheckTimeLimit() bool {
	if qi.Definition.TimeLimit == 0 {
		return false
	}
	
	if time.Since(qi.AcceptedAt) > qi.Definition.TimeLimit {
		qi.Failed = true
		now := time.Now()
		qi.ExpiredAt = &now
		return true
	}
	return false
}

// Complete marks quest as complete
func (qi *QuestInstance) Complete() {
	if !qi.Failed && !qi.Completed {
		qi.Completed = true
		now := time.Now()
		qi.CompletedAt = &now
	}
}

// QuestRegistry holds all quest definitions
var QuestRegistry = make(map[string]*QuestDefinition)

// RegisterQuest registers a quest
func RegisterQuest(quest *QuestDefinition) {
	QuestRegistry[quest.ID] = quest
}

func init() {
	// Starter quests
	RegisterQuest(&QuestDefinition{
		ID: "starter_1", Name: "Getting Started",
		Description: "Gather 10 dirt blocks",
		Type: QUEST_GATHER,
		Objectives: []QuestObjective{{Type: "gather", Target: "dirt", Quantity: 10}},
		Reward: QuestReward{XP: 50, Gold: 10},
	})
	RegisterQuest(&QuestDefinition{
		ID: "starter_2", Name: "First Craft",
		Description: "Craft a workbench",
		Type: QUEST_CRAFT,
		Objectives: []QuestObjective{{Type: "craft", Target: "workbench", Quantity: 1}},
		Reward: QuestReward{XP: 100, Gold: 25},
		Prereqs: []string{"starter_1"},
	})
	RegisterQuest(&QuestDefinition{
		ID: "starter_3", Name: "Tool Up",
		Description: "Craft a wooden pickaxe",
		Type: QUEST_CRAFT,
		Objectives: []QuestObjective{{Type: "craft", Target: "wooden_pickaxe", Quantity: 1}},
		Reward: QuestReward{XP: 75, Gold: 15, Items: map[string]int{"coal": 5}},
		Prereqs: []string{"starter_2"},
	})
	
	// Mining quests
	RegisterQuest(&QuestDefinition{
		ID: "miner_1", Name: "Stone Miner",
		Description: "Mine 50 stone blocks",
		Type: QUEST_GATHER,
		Objectives: []QuestObjective{{Type: "gather", Target: "stone", Quantity: 50}},
		Reward: QuestReward{XP: 150, Gold: 30},
	})
	RegisterQuest(&QuestDefinition{
		ID: "miner_2", Name: "Deep Delver",
		Description: "Reach the middle layer",
		Type: QUEST_EXPLORE,
		Objectives: []QuestObjective{{Type: "explore", Target: "layer_1", Quantity: 1}},
		Reward: QuestReward{XP: 200, Gold: 40},
		Prereqs: []string{"miner_1"},
	})
	
	// Survival quests
	RegisterQuest(&QuestDefinition{
		ID: "survival_1", Name: "First Night",
		Description: "Survive your first night",
		Type: QUEST_SURVIVE,
		Objectives: []QuestObjective{{Type: "survive", Target: "night", Quantity: 1}},
		Reward: QuestReward{XP: 100, Gold: 20},
		TimeLimit: 15 * time.Minute,
	})
	
	// Combat quests
	RegisterQuest(&QuestDefinition{
		ID: "combat_1", Name: "First Blood",
		Description: "Defeat 3 enemies",
		Type: QUEST_COMBAT,
		Objectives: []QuestObjective{{Type: "kill", Target: "any", Quantity: 3}},
		Reward: QuestReward{XP: 150, Gold: 35, Items: map[string]int{"healing_potion": 2}},
	})
}
