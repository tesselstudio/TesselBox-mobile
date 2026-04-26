package achievements

import (
	"time"
)

// AchievementCategory represents achievement categories
type AchievementCategory int

const (
	CATEGORY_GENERAL AchievementCategory = iota
	CATEGORY_MINING
	CATEGORY_CRAFTING
	CATEGORY_SURVIVAL
	CATEGORY_EXPLORATION
	CATEGORY_COMBAT
	CATEGORY_BUILDING
)

func (ac AchievementCategory) String() string {
	switch ac {
	case CATEGORY_GENERAL:
		return "General"
	case CATEGORY_MINING:
		return "Mining"
	case CATEGORY_CRAFTING:
		return "Crafting"
	case CATEGORY_SURVIVAL:
		return "Survival"
	case CATEGORY_EXPLORATION:
		return "Exploration"
	case CATEGORY_COMBAT:
		return "Combat"
	case CATEGORY_BUILDING:
		return "Building"
	default:
		return "Unknown"
	}
}

// AchievementTier represents rarity tiers
type AchievementTier int

const (
	TIER_BRONZE AchievementTier = iota
	TIER_SILVER
	TIER_GOLD
	TIER_PLATINUM
	TIER_DIAMOND
)

func (at AchievementTier) String() string {
	switch at {
	case TIER_BRONZE:
		return "Bronze"
	case TIER_SILVER:
		return "Silver"
	case TIER_GOLD:
		return "Gold"
	case TIER_PLATINUM:
		return "Platinum"
	case TIER_DIAMOND:
		return "Diamond"
	default:
		return "Unknown"
	}
}

// AchievementDefinition defines an achievement
type AchievementDefinition struct {
	ID          string
	Name        string
	Description string
	Category    AchievementCategory
	Tier        AchievementTier
	Icon        string
	Hidden      bool // Secret achievement
	MaxProgress int  // For progress-based achievements
	RewardXP    int
	RewardItem  string // Item ID if any
}

// IsProgressBased returns true if achievement has progress tracking
func (ad *AchievementDefinition) IsProgressBased() bool {
	return ad.MaxProgress > 1
}

// AchievementProgress tracks progress for an achievement
type AchievementProgress struct {
	Definition  *AchievementDefinition
	Unlocked    bool
	Progress    int
	UnlockedAt  *time.Time
}

// IsComplete returns true if progress-based achievement is complete
func (ap *AchievementProgress) IsComplete() bool {
	if !ap.Definition.IsProgressBased() {
		return ap.Unlocked
	}
	return ap.Progress >= ap.Definition.MaxProgress
}

// GetProgressPercentage returns 0-100 completion
func (ap *AchievementProgress) GetProgressPercentage() float64 {
	if !ap.Definition.IsProgressBased() {
		if ap.Unlocked {
			return 100.0
		}
		return 0.0
	}
	return float64(ap.Progress) / float64(ap.Definition.MaxProgress) * 100.0
}

// AchievementRegistry holds all achievement definitions
var AchievementRegistry = make(map[string]*AchievementDefinition)

// RegisterAchievement registers an achievement definition
func RegisterAchievement(def *AchievementDefinition) {
	AchievementRegistry[def.ID] = def
}

// GetAchievementDefinition retrieves an achievement by ID
func GetAchievementDefinition(id string) *AchievementDefinition {
	return AchievementRegistry[id]
}

// GetAchievementsByCategory returns all achievements in a category
func GetAchievementsByCategory(category AchievementCategory) []*AchievementDefinition {
	var result []*AchievementDefinition
	for _, def := range AchievementRegistry {
		if def.Category == category {
			result = append(result, def)
		}
	}
	return result
}

func init() {
	// Register Mining achievements
	RegisterAchievement(&AchievementDefinition{
		ID:          "miner_bronze",
		Name:        "Novice Miner",
		Description: "Mine 10 blocks",
		Category:    CATEGORY_MINING,
		Tier:        TIER_BRONZE,
		MaxProgress: 10,
		RewardXP:    50,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "miner_silver",
		Name:        "Experienced Miner",
		Description: "Mine 100 blocks",
		Category:    CATEGORY_MINING,
		Tier:        TIER_SILVER,
		MaxProgress: 100,
		RewardXP:    200,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "miner_gold",
		Name:        "Master Miner",
		Description: "Mine 1000 blocks",
		Category:    CATEGORY_MINING,
		Tier:        TIER_GOLD,
		MaxProgress: 1000,
		RewardXP:    1000,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "diamond_hunter",
		Name:        "Diamond Hunter",
		Description: "Find your first diamond",
		Category:    CATEGORY_MINING,
		Tier:        TIER_PLATINUM,
		RewardXP:    500,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "deep_diver",
		Name:        "Deep Diver",
		Description: "Reach the deepest layer",
		Category:    CATEGORY_MINING,
		Tier:        TIER_PLATINUM,
		RewardXP:    300,
	})

	// Register Crafting achievements
	RegisterAchievement(&AchievementDefinition{
		ID:          "crafter_bronze",
		Name:        "Novice Crafter",
		Description: "Craft 10 items",
		Category:    CATEGORY_CRAFTING,
		Tier:        TIER_BRONZE,
		MaxProgress: 10,
		RewardXP:    50,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "crafter_silver",
		Name:        "Experienced Crafter",
		Description: "Craft 50 items",
		Category:    CATEGORY_CRAFTING,
		Tier:        TIER_SILVER,
		MaxProgress: 50,
		RewardXP:    200,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "tool_smith",
		Name:        "Tool Smith",
		Description: "Craft all types of pickaxes",
		Category:    CATEGORY_CRAFTING,
		Tier:        TIER_GOLD,
		RewardXP:    500,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "armor_master",
		Name:        "Armor Master",
		Description: "Craft a full set of armor",
		Category:    CATEGORY_CRAFTING,
		Tier:        TIER_PLATINUM,
		RewardXP:    750,
	})

	// Register Survival achievements
	RegisterAchievement(&AchievementDefinition{
		ID:          "first_night",
		Name:        "First Night",
		Description: "Survive your first night",
		Category:    CATEGORY_SURVIVAL,
		Tier:        TIER_BRONZE,
		RewardXP:    100,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "week_warrior",
		Name:        "Week Warrior",
		Description: "Survive for 7 days",
		Category:    CATEGORY_SURVIVAL,
		Tier:        TIER_SILVER,
		MaxProgress: 7,
		RewardXP:    500,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "month_master",
		Name:        "Month Master",
		Description: "Survive for 30 days",
		Category:    CATEGORY_SURVIVAL,
		Tier:        TIER_GOLD,
		MaxProgress: 30,
		RewardXP:    2000,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "iron_gut",
		Name:        "Iron Gut",
		Description: "Eat 100 food items",
		Category:    CATEGORY_SURVIVAL,
		Tier:        TIER_SILVER,
		MaxProgress: 100,
		RewardXP:    200,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "hydrated",
		Name:        "Hydrated",
		Description: "Drink 50 beverages",
		Category:    CATEGORY_SURVIVAL,
		Tier:        TIER_SILVER,
		MaxProgress: 50,
		RewardXP:    150,
	})

	// Register Exploration achievements
	RegisterAchievement(&AchievementDefinition{
		ID:          "explorer_bronze",
		Name:        "Novice Explorer",
		Description: "Travel 1000 blocks",
		Category:    CATEGORY_EXPLORATION,
		Tier:        TIER_BRONZE,
		MaxProgress: 1000,
		RewardXP:    100,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "explorer_silver",
		Name:        "Seasoned Explorer",
		Description: "Travel 10000 blocks",
		Category:    CATEGORY_EXPLORATION,
		Tier:        TIER_SILVER,
		MaxProgress: 10000,
		RewardXP:    500,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "world_wanderer",
		Name:        "World Wanderer",
		Description: "Travel 100000 blocks",
		Category:    CATEGORY_EXPLORATION,
		Tier:        TIER_GOLD,
		MaxProgress: 100000,
		RewardXP:    2500,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "portal_traveler",
		Name:        "Portal Traveler",
		Description: "Use a portal to another dimension",
		Category:    CATEGORY_EXPLORATION,
		Tier:        TIER_PLATINUM,
		RewardXP:    1000,
	})

	// Register Combat achievements
	RegisterAchievement(&AchievementDefinition{
		ID:          "first_blood",
		Name:        "First Blood",
		Description: "Defeat your first enemy",
		Category:    CATEGORY_COMBAT,
		Tier:        TIER_BRONZE,
		RewardXP:    100,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "monster_hunter",
		Name:        "Monster Hunter",
		Description: "Defeat 10 enemies",
		Category:    CATEGORY_COMBAT,
		Tier:        TIER_SILVER,
		MaxProgress: 10,
		RewardXP:    300,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "slayer",
		Name:        "Slayer",
		Description: "Defeat 50 enemies",
		Category:    CATEGORY_COMBAT,
		Tier:        TIER_GOLD,
		MaxProgress: 50,
		RewardXP:    1000,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "untouchable",
		Name:        "Untouchable",
		Description: "Defeat an enemy without taking damage",
		Category:    CATEGORY_COMBAT,
		Tier:        TIER_PLATINUM,
		RewardXP:    500,
	})

	// Register Building achievements
	RegisterAchievement(&AchievementDefinition{
		ID:          "builder_bronze",
		Name:        "Novice Builder",
		Description: "Place 50 blocks",
		Category:    CATEGORY_BUILDING,
		Tier:        TIER_BRONZE,
		MaxProgress: 50,
		RewardXP:    50,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "builder_silver",
		Name:        "Builder",
		Description: "Place 500 blocks",
		Category:    CATEGORY_BUILDING,
		Tier:        TIER_SILVER,
		MaxProgress: 500,
		RewardXP:    300,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "architect",
		Name:        "Architect",
		Description: "Place 5000 blocks",
		Category:    CATEGORY_BUILDING,
		Tier:        TIER_GOLD,
		MaxProgress: 5000,
		RewardXP:    1500,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "home_sweet_home",
		Name:        "Home Sweet Home",
		Description: "Build a shelter with a bed, chest, and crafting station",
		Category:    CATEGORY_BUILDING,
		Tier:        TIER_SILVER,
		RewardXP:    400,
	})

	// Register General achievements
	RegisterAchievement(&AchievementDefinition{
		ID:          "getting_started",
		Name:        "Getting Started",
		Description: "Play for the first time",
		Category:    CATEGORY_GENERAL,
		Tier:        TIER_BRONZE,
		RewardXP:    25,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "dedicated",
		Name:        "Dedicated",
		Description: "Play for 1 hour",
		Category:    CATEGORY_GENERAL,
		Tier:        TIER_SILVER,
		RewardXP:    200,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "committed",
		Name:        "Committed",
		Description: "Play for 10 hours",
		Category:    CATEGORY_GENERAL,
		Tier:        TIER_GOLD,
		RewardXP:    1000,
	})
	RegisterAchievement(&AchievementDefinition{
		ID:          "completionist",
		Name:        "Completionist",
		Description: "Unlock all achievements",
		Category:    CATEGORY_GENERAL,
		Tier:        TIER_DIAMOND,
		Hidden:      true,
		RewardXP:    5000,
	})
}
