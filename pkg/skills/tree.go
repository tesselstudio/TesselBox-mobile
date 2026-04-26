package skills

import (
	"time"
)

// SkillTree represents different skill categories
type SkillTree int

const (
	TREE_MINING SkillTree = iota
	TREE_CRAFTING
	TREE_COMBAT
	TREE_SURVIVAL
	TREE_MOVEMENT
)

func (st SkillTree) String() string {
	switch st {
	case TREE_MINING:
		return "Mining"
	case TREE_CRAFTING:
		return "Crafting"
	case TREE_COMBAT:
		return "Combat"
	case TREE_SURVIVAL:
		return "Survival"
	case TREE_MOVEMENT:
		return "Movement"
	default:
		return "Unknown"
	}
}

// SkillTier represents skill node tiers
type SkillTier int

const (
	SKILL_TIER_1 SkillTier = iota // Basic
	SKILL_TIER_2                  // Advanced
	SKILL_TIER_3                  // Expert
	SKILL_TIER_4                  // Master
)

// SkillEffect represents what a skill does
type SkillEffect struct {
	Type     string
	Target   string
	Modifier float64
}

// SkillNode represents a skill in the tree
type SkillNode struct {
	ID          string
	Tree        SkillTree
	Name        string
	Description string
	Tier        SkillTier
	X, Y        int // Position in tree UI
	Cost        int // XP Cost
	Icon        string
	Effects     []SkillEffect
	Requires    []string // Required skill IDs
	MaxLevel    int
}

// SkillInstance represents player progress in a skill
type SkillInstance struct {
	Definition *SkillNode
	Unlocked   bool
	Level      int
	UnlockedAt *time.Time
}

// GetCurrentEffect returns the effect at current level
func (si *SkillInstance) GetCurrentEffect() float64 {
	if !si.Unlocked || si.Level == 0 {
		return 0
	}
	
	totalEffect := 0.0
	for _, effect := range si.Definition.Effects {
		totalEffect += effect.Modifier * float64(si.Level)
	}
	return totalEffect
}

// CanUnlock checks if skill can be unlocked
func (si *SkillInstance) CanUnlock(unlockedSkills map[string]*SkillInstance) bool {
	if si.Unlocked {
		return false
	}
	
	for _, reqID := range si.Definition.Requires {
		req, exists := unlockedSkills[reqID]
		if !exists || !req.Unlocked {
			return false
		}
	}
	return true
}

// SkillRegistry holds all skill definitions
var SkillRegistry = make(map[string]*SkillNode)

// RegisterSkill registers a skill
func RegisterSkill(skill *SkillNode) {
	SkillRegistry[skill.ID] = skill
}

func init() {
	// Mining tree
	RegisterSkill(&SkillNode{
		ID: "mining_speed_1", Tree: TREE_MINING, Tier: SKILL_TIER_1,
		Name: "Efficient Mining", Description: "Mine 10% faster",
		Cost: 100, Effects: []SkillEffect{{Type: "speed", Target: "mining", Modifier: 0.10}},
		MaxLevel: 1,
	})
	RegisterSkill(&SkillNode{
		ID: "mining_speed_2", Tree: TREE_MINING, Tier: SKILL_TIER_2,
		Name: "Master Miner", Description: "Mine 20% faster",
		Cost: 300, Effects: []SkillEffect{{Type: "speed", Target: "mining", Modifier: 0.20}},
		Requires: []string{"mining_speed_1"}, MaxLevel: 1,
	})
	RegisterSkill(&SkillNode{
		ID: "mining_fortune", Tree: TREE_MINING, Tier: SKILL_TIER_3,
		Name: "Mining Fortune", Description: "Chance for double drops",
		Cost: 500, Effects: []SkillEffect{{Type: "fortune", Target: "drops", Modifier: 0.25}},
		Requires: []string{"mining_speed_2"}, MaxLevel: 1,
	})
	
	// Crafting tree
	RegisterSkill(&SkillNode{
		ID: "crafting_speed", Tree: TREE_CRAFTING, Tier: SKILL_TIER_1,
		Name: "Fast Crafting", Description: "Craft 15% faster",
		Cost: 100, Effects: []SkillEffect{{Type: "speed", Target: "crafting", Modifier: 0.15}},
		MaxLevel: 1,
	})
	RegisterSkill(&SkillNode{
		ID: "crafting_efficiency", Tree: TREE_CRAFTING, Tier: SKILL_TIER_2,
		Name: "Material Efficiency", Description: "Chance to save materials",
		Cost: 250, Effects: []SkillEffect{{Type: "save", Target: "materials", Modifier: 0.20}},
		Requires: []string{"crafting_speed"}, MaxLevel: 1,
	})
	
	// Combat tree
	RegisterSkill(&SkillNode{
		ID: "combat_damage", Tree: TREE_COMBAT, Tier: SKILL_TIER_1,
		Name: "Sharp Strikes", Description: "Deal 10% more damage",
		Cost: 100, Effects: []SkillEffect{{Type: "damage", Target: "melee", Modifier: 0.10}},
		MaxLevel: 1,
	})
	RegisterSkill(&SkillNode{
		ID: "combat_defense", Tree: TREE_COMBAT, Tier: SKILL_TIER_1,
		Name: "Tough Skin", Description: "Take 10% less damage",
		Cost: 100, Effects: []SkillEffect{{Type: "defense", Target: "all", Modifier: 0.10}},
		MaxLevel: 1,
	})
	
	// Survival tree
	RegisterSkill(&SkillNode{
		ID: "survival_hunger", Tree: TREE_SURVIVAL, Tier: SKILL_TIER_1,
		Name: "Slow Metabolism", Description: "Hunger drains 15% slower",
		Cost: 100, Effects: []SkillEffect{{Type: "resist", Target: "hunger", Modifier: 0.15}},
		MaxLevel: 1,
	})
	RegisterSkill(&SkillNode{
		ID: "survival_health", Tree: TREE_SURVIVAL, Tier: SKILL_TIER_1,
		Name: "Regeneration", Description: "Slow health regeneration",
		Cost: 150, Effects: []SkillEffect{{Type: "regen", Target: "health", Modifier: 0.05}},
		MaxLevel: 1,
	})
	
	// Movement tree
	RegisterSkill(&SkillNode{
		ID: "movement_speed", Tree: TREE_MOVEMENT, Tier: SKILL_TIER_1,
		Name: "Swift Foot", Description: "Move 10% faster",
		Cost: 100, Effects: []SkillEffect{{Type: "speed", Target: "movement", Modifier: 0.10}},
		MaxLevel: 1,
	})
}
