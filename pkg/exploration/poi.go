package exploration

import (
	"time"
)

// POIType represents different types of points of interest
type POIType int

const (
	POI_RUINS POIType = iota
	POI_CAVE
	POI_FLOATING_ISLAND
	POI_ANCIENT_STRUCTURE
	POI_UNDERGROUND_TEMPLE
	POI_ABANDONED_MINE
	POI_MONOLITH
	POI_PORTAL
	POI_LOOT_CACHE
	POI_LANDMARK
)

func (pt POIType) String() string {
	switch pt {
	case POI_RUINS:
		return "Ancient Ruins"
	case POI_CAVE:
		return "Mysterious Cave"
	case POI_FLOATING_ISLAND:
		return "Floating Island"
	case POI_ANCIENT_STRUCTURE:
		return "Ancient Structure"
	case POI_UNDERGROUND_TEMPLE:
		return "Underground Temple"
	case POI_ABANDONED_MINE:
		return "Abandoned Mine"
	case POI_MONOLITH:
		return "Strange Monolith"
	case POI_PORTAL:
		return "Mystical Portal"
	case POI_LOOT_CACHE:
		return "Hidden Loot Cache"
	case POI_LANDMARK:
		return "Notable Landmark"
	default:
		return "Unknown"
	}
}

// POIRarity represents rarity tiers
type POIRarity int

const (
	RARITY_COMMON POIRarity = iota
	RARITY_UNCOMMON
	RARITY_RARE
	RARITY_EPIC
	RARITY_LEGENDARY
)

func (pr POIRarity) String() string {
	switch pr {
	case RARITY_COMMON:
		return "Common"
	case RARITY_UNCOMMON:
		return "Uncommon"
	case RARITY_RARE:
		return "Rare"
	case RARITY_EPIC:
		return "Epic"
	case RARITY_LEGENDARY:
		return "Legendary"
	default:
		return "Unknown"
	}
}

// Reward represents loot from a POI
type Reward struct {
	ItemID   string
	Quantity int
	Chance   float64 // 0-1 probability
}

// PointOfInterest represents a discoverable location
type PointOfInterest struct {
	ID          string
	Type        POIType
	Name        string
	Description string
	Rarity      POIRarity
	X, Y        float64 // World coordinates
	Layer       int     // World layer
	Discovered  bool
	DiscoveredAt *time.Time
	Explored    bool
	ExploredAt  *time.Time
	Rewards     []Reward
	XPValue     int
	Range       float64 // Detection range
	CustomData  map[string]interface{}
}

// CanDiscover checks if player can discover this POI
func (poi *PointOfInterest) CanDiscover(playerX, playerY float64, playerLayer int, detectionRange float64) bool {
	if poi.Discovered {
		return false
	}
	if playerLayer != poi.Layer {
		return false
	}
	
	dx := poi.X - playerX
	dy := poi.Y - playerY
	distance := dx*dx + dy*dy
	
	// Use POI's own range or player's detection range, whichever is larger
	effectiveRange := poi.Range
	if detectionRange > effectiveRange {
		effectiveRange = detectionRange
	}
	
	return distance <= effectiveRange*effectiveRange
}

// GetDistance returns distance from a point
func (poi *PointOfInterest) GetDistance(x, y float64) float64 {
	dx := poi.X - x
	dy := poi.Y - y
	return dx*dx + dy*dy
}

// Discover marks the POI as discovered
func (poi *PointOfInterest) Discover() {
	if !poi.Discovered {
		poi.Discovered = true
		now := time.Now()
		poi.DiscoveredAt = &now
	}
}

// Explore marks the POI as explored and returns rewards
func (poi *PointOfInterest) Explore() (int, []Reward) {
	if !poi.Explored {
		poi.Explored = true
		now := time.Now()
		poi.ExploredAt = &now
		
		// Determine actual rewards based on chance
		var actualRewards []Reward
		for _, reward := range poi.Rewards {
			if reward.Chance >= 1.0 || randFloat() < reward.Chance {
				actualRewards = append(actualRewards, reward)
			}
		}
		
		return poi.XPValue, actualRewards
	}
	return 0, nil
}

// POIRegistry holds POI definitions
var POIRegistry = make(map[POIType]*PointOfInterest)

// RegisterPOI registers a POI template
func RegisterPOI(poi *PointOfInterest) {
	POIRegistry[poi.Type] = poi
}

// randFloat returns random float 0-1
func randFloat() float64 {
	return float64(time.Now().UnixNano()%1000) / 1000.0
}

func init() {
	// Register POI templates
	RegisterPOI(&PointOfInterest{
		Type:        POI_RUINS,
		Name:        "Ancient Ruins",
		Description: "Remnants of a forgotten civilization",
		Rarity:      RARITY_COMMON,
		Range:       150,
		XPValue:     100,
		Rewards: []Reward{
			{ItemID: "stone_block", Quantity: 10, Chance: 1.0},
			{ItemID: "workbench", Quantity: 1, Chance: 0.3},
		},
	})

	RegisterPOI(&PointOfInterest{
		Type:        POI_CAVE,
		Name:        "Mysterious Cave",
		Description: "A dark cave with unknown depths",
		Rarity:      RARITY_COMMON,
		Range:       100,
		XPValue:     75,
		Rewards: []Reward{
			{ItemID: "coal", Quantity: 5, Chance: 0.8},
			{ItemID: "iron_ore", Quantity: 3, Chance: 0.4},
		},
	})

	RegisterPOI(&PointOfInterest{
		Type:        POI_FLOATING_ISLAND,
		Name:        "Floating Island",
		Description: "A mysterious island floating in the sky",
		Rarity:      RARITY_EPIC,
		Range:       300,
		XPValue:     500,
		Rewards: []Reward{
			{ItemID: "diamond", Quantity: 2, Chance: 0.5},
			{ItemID: "gold_ingot", Quantity: 5, Chance: 0.7},
			{ItemID: "wings", Quantity: 1, Chance: 0.2},
		},
	})

	RegisterPOI(&PointOfInterest{
		Type:        POI_ANCIENT_STRUCTURE,
		Name:        "Ancient Structure",
		Description: "A mysterious building of unknown origin",
		Rarity:      RARITY_RARE,
		Range:       200,
		XPValue:     250,
		Rewards: []Reward{
			{ItemID: "iron_ingot", Quantity: 5, Chance: 0.6},
			{ItemID: "anvil", Quantity: 1, Chance: 0.3},
		},
	})

	RegisterPOI(&PointOfInterest{
		Type:        POI_LOOT_CACHE,
		Name:        "Hidden Loot Cache",
		Description: "A hidden cache of valuable items",
		Rarity:      RARITY_UNCOMMON,
		Range:       80,
		XPValue:     150,
		Rewards: []Reward{
			{ItemID: "wooden_pickaxe", Quantity: 1, Chance: 0.5},
			{ItemID: "coal", Quantity: 10, Chance: 0.8},
			{ItemID: "iron_ingot", Quantity: 2, Chance: 0.3},
		},
	})

	RegisterPOI(&PointOfInterest{
		Type:        POI_PORTAL,
		Name:        "Mystical Portal",
		Description: "A portal to another dimension",
		Rarity:      RARITY_LEGENDARY,
		Range:       250,
		XPValue:     1000,
		Rewards: []Reward{
			{ItemID: "randomland_portal", Quantity: 1, Chance: 1.0},
			{ItemID: "diamond", Quantity: 5, Chance: 0.5},
		},
	})
}
