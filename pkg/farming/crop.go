package farming

import (
	"time"
)

// CropType represents different crop types
type CropType int

const (
	CROP_WHEAT CropType = iota
	CROP_CARROT
	CROP_POTATO
	CROP_CORN
	CROP_TOMATO
	CROP_TREE_OAK
	CROP_TREE_PINE
	CROP_TREE_BIRCH
)

func (ct CropType) String() string {
	switch ct {
	case CROP_WHEAT:
		return "Wheat"
	case CROP_CARROT:
		return "Carrot"
	case CROP_POTATO:
		return "Potato"
	case CROP_CORN:
		return "Corn"
	case CROP_TOMATO:
		return "Tomato"
	case CROP_TREE_OAK:
		return "Oak Tree"
	case CROP_TREE_PINE:
		return "Pine Tree"
	case CROP_TREE_BIRCH:
		return "Birch Tree"
	default:
		return "Unknown"
	}
}

// GrowthStage represents crop growth stages
type GrowthStage int

const (
	STAGE_SEED GrowthStage = iota
	STAGE_SPROUT
	STAGE_GROWING
	STAGE_MATURE
	STAGE_HARVESTABLE
)

func (gs GrowthStage) String() string {
	switch gs {
	case STAGE_SEED:
		return "Seed"
	case STAGE_SPROUT:
		return "Sprout"
	case STAGE_GROWING:
		return "Growing"
	case STAGE_MATURE:
		return "Mature"
	case STAGE_HARVESTABLE:
		return "Ready to Harvest"
	default:
		return "Unknown"
	}
}

// CropYield represents what a crop produces when harvested
type CropYield struct {
	ItemID   string
	Quantity int
	Chance   float64
}

// CropDefinition defines a crop type
type CropDefinition struct {
	Type           CropType
	Name           string
	Description    string
	GrowthTime     time.Duration // Time to reach harvestable
	WaterNeeded    int           // Water units needed per growth stage
	Stages         int           // Number of growth stages
	Yields         []CropYield
	XPValue        int
	IsTree         bool
	RegrowAfter    time.Duration // For multi-harvest crops
	Seasons        []string        // Best growing seasons
}

// CropInstance represents a planted crop
type CropInstance struct {
	ID            string
	Definition    *CropDefinition
	X, Y          float64
	Layer         int
	PlantedAt     time.Time
	Stage         GrowthStage
	WaterLevel    int
	Health        float64 // 0-100
	LastWatered   time.Time
	LastHarvested *time.Time
	IsDead        bool
	CustomData    map[string]interface{}
}

// GetGrowthProgress returns progress to next stage or harvest (0-1)
func (ci *CropInstance) GetGrowthProgress() float64 {
	if ci.IsDead || ci.Stage == STAGE_HARVESTABLE {
		return 1.0
	}

	stageDuration := ci.Definition.GrowthTime / time.Duration(ci.Definition.Stages)
	stageStart := ci.PlantedAt.Add(stageDuration * time.Duration(ci.Stage))
	elapsed := time.Since(stageStart)

	progress := float64(elapsed) / float64(stageDuration)
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// GetTimeUntilHarvest returns time until harvestable
func (ci *CropInstance) GetTimeUntilHarvest() time.Duration {
	if ci.Stage == STAGE_HARVESTABLE {
		return 0
	}

	remainingStages := ci.Definition.Stages - int(ci.Stage)
	stageDuration := ci.Definition.GrowthTime / time.Duration(ci.Definition.Stages)

	return stageDuration * time.Duration(remainingStages)
}

// NeedsWater returns true if crop needs water
func (ci *CropInstance) NeedsWater() bool {
	return ci.WaterLevel < ci.Definition.WaterNeeded
}

// Water adds water to the crop
func (ci *CropInstance) Water(amount int) {
	ci.WaterLevel += amount
	ci.LastWatered = time.Now()
	if ci.WaterLevel > ci.Definition.WaterNeeded*2 {
		ci.WaterLevel = ci.Definition.WaterNeeded * 2
	}
}

// CanHarvest returns true if crop is ready to harvest
func (ci *CropInstance) CanHarvest() bool {
	return ci.Stage == STAGE_HARVESTABLE && !ci.IsDead
}

// Harvest harvests the crop and returns yields
func (ci *CropInstance) Harvest() []CropYield {
	if !ci.CanHarvest() {
		return nil
	}

	var yields []CropYield
	for _, yield := range ci.Definition.Yields {
		if yield.Chance >= 1.0 || randFloat() < yield.Chance {
			yields = append(yields, yield)
		}
	}

	if ci.Definition.IsTree && ci.Definition.RegrowAfter > 0 {
		// Trees regrow
		ci.Stage = STAGE_MATURE
		now := time.Now()
		ci.LastHarvested = &now
	} else {
		// One-time harvest
		ci.IsDead = true
	}

	return yields
}

// UpdateGrowth updates crop growth based on conditions
func (ci *CropInstance) UpdateGrowth(deltaTime float64, dayTime float64) {
	if ci.IsDead || ci.Stage == STAGE_HARVESTABLE {
		return
	}

	// Check water
	if ci.WaterLevel <= 0 {
		// Crop loses health without water
		ci.Health -= float64(deltaTime) * 5
		if ci.Health <= 0 {
			ci.IsDead = true
			return
		}
	}

	// Growth requires water
	if ci.WaterLevel >= ci.Definition.WaterNeeded {
		progress := ci.GetGrowthProgress()
		if progress >= 1.0 {
			ci.advanceStage()
		}
	}

	// Consume water slowly
	if deltaTime > 60 { // Every minute
		ci.WaterLevel--
		if ci.WaterLevel < 0 {
			ci.WaterLevel = 0
		}
	}
}

// advanceStage moves to next growth stage
func (ci *CropInstance) advanceStage() {
	if ci.Stage < STAGE_HARVESTABLE {
		ci.Stage++
		ci.Health = 100
	}
}

// CropRegistry holds crop definitions
var CropRegistry = make(map[CropType]*CropDefinition)

// RegisterCrop registers a crop definition
func RegisterCrop(def *CropDefinition) {
	CropRegistry[def.Type] = def
}

// randFloat returns random float 0-1
func randFloat() float64 {
	return float64(time.Now().UnixNano()%1000) / 1000.0
}

func init() {
	// Register Wheat
	RegisterCrop(&CropDefinition{
		Type:        CROP_WHEAT,
		Name:        "Wheat",
		Description: "Basic grain crop",
		GrowthTime:  5 * time.Minute,
		WaterNeeded: 1,
		Stages:      4,
		XPValue:     10,
		Yields: []CropYield{
			{ItemID: "wheat", Quantity: 2, Chance: 1.0},
			{ItemID: "wheat_seeds", Quantity: 1, Chance: 0.8},
		},
	})

	// Register Carrot
	RegisterCrop(&CropDefinition{
		Type:        CROP_CARROT,
		Name:        "Carrot",
		Description: "Quick growing root vegetable",
		GrowthTime:  3 * time.Minute,
		WaterNeeded: 1,
		Stages:      3,
		XPValue:     8,
		Yields: []CropYield{
			{ItemID: "carrot", Quantity: 2, Chance: 1.0},
			{ItemID: "carrot_seeds", Quantity: 1, Chance: 0.7},
		},
	})

	// Register Potato
	RegisterCrop(&CropDefinition{
		Type:        CROP_POTATO,
		Name:        "Potato",
		Description: "Versatile underground crop",
		GrowthTime:  6 * time.Minute,
		WaterNeeded: 2,
		Stages:      4,
		XPValue:     12,
		Yields: []CropYield{
			{ItemID: "potato", Quantity: 2, Chance: 1.0},
			{ItemID: "potato", Quantity: 1, Chance: 0.5},
		},
	})

	// Register Corn
	RegisterCrop(&CropDefinition{
		Type:        CROP_CORN,
		Name:        "Corn",
		Description: "Tall grain crop",
		GrowthTime:  8 * time.Minute,
		WaterNeeded: 2,
		Stages:      5,
		XPValue:     15,
		Yields: []CropYield{
			{ItemID: "corn", Quantity: 2, Chance: 1.0},
			{ItemID: "corn_seeds", Quantity: 1, Chance: 0.8},
		},
	})

	// Register Trees
	RegisterCrop(&CropDefinition{
		Type:        CROP_TREE_OAK,
		Name:        "Oak Tree",
		Description: "Sturdy hardwood tree",
		GrowthTime:  15 * time.Minute,
		WaterNeeded: 3,
		Stages:      5,
		XPValue:     30,
		IsTree:      true,
		RegrowAfter: 10 * time.Minute,
		Yields: []CropYield{
			{ItemID: "log", Quantity: 4, Chance: 1.0},
			{ItemID: "sapling", Quantity: 1, Chance: 0.3},
		},
	})
}
