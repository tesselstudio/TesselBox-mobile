package basebuilding

import (
	"time"
)

// StructureType represents building types
type StructureType int

const (
	STRUCT_WALL StructureType = iota
	STRUCT_DOOR
	STRUCT_TRAP
	STRUCT_TURRET
	STRUCT_CHEST
	STRUCT_BED
	STRUCT_LAMP
	STRUCT_DECORATION
)

func (st StructureType) String() string {
	switch st {
	case STRUCT_WALL:
		return "Wall"
	case STRUCT_DOOR:
		return "Door"
	case STRUCT_TRAP:
		return "Trap"
	case STRUCT_TURRET:
		return "Turret"
	case STRUCT_CHEST:
		return "Chest"
	case STRUCT_BED:
		return "Bed"
	case STRUCT_LAMP:
		return "Lamp"
	case STRUCT_DECORATION:
		return "Decoration"
	default:
		return "Unknown"
	}
}

// StructureDef defines a structure
type StructureDef struct {
	Type          StructureType
	Name          string
	Description   string
	Health        int
	Defense       int
	Cost          map[string]int // Item ID -> quantity
	XPValue       int
	IsSolid       bool
	IsInteractive bool
}

// StructureInstance represents a built structure
type StructureInstance struct {
	ID       string
	Def      *StructureDef
	X, Y     float64
	Layer    int
	Health   int
	BuiltAt  time.Time
	BuiltBy  string
	IsActive bool
}

// TakeDamage applies damage to structure
func (si *StructureInstance) TakeDamage(damage int) {
	si.Health -= damage
	if si.Health < 0 {
		si.Health = 0
	}
}

// IsDestroyed returns true if health is 0
func (si *StructureInstance) IsDestroyed() bool {
	return si.Health <= 0
}

// Repair repairs the structure
func (si *StructureInstance) Repair(amount int) {
	si.Health += amount
	if si.Health > si.Def.Health {
		si.Health = si.Def.Health
	}
}

// StructureRegistry holds structure definitions
var StructureRegistry = make(map[StructureType]*StructureDef)

// RegisterStructure registers a structure
func RegisterStructure(def *StructureDef) {
	StructureRegistry[def.Type] = def
}

func init() {
	// Register structures
	RegisterStructure(&StructureDef{
		Type:        STRUCT_WALL,
		Name:        "Stone Wall",
		Description: "A sturdy defensive wall",
		Health:      100,
		Defense:     5,
		IsSolid:     true,
		Cost:        map[string]int{"stone": 5},
		XPValue:     10,
	})
	RegisterStructure(&StructureDef{
		Type:          STRUCT_DOOR,
		Name:          "Wooden Door",
		Description:   "A door for your shelter",
		Health:        50,
		Defense:       2,
		IsSolid:       false,
		IsInteractive: true,
		Cost:          map[string]int{"wood": 5},
		XPValue:       5,
	})
	RegisterStructure(&StructureDef{
		Type:          STRUCT_TRAP,
		Name:          "Spike Trap",
		Description:   "Damages enemies that step on it",
		Health:        30,
		Defense:       0,
		IsSolid:       false,
		IsInteractive: true,
		Cost:          map[string]int{"wood": 3, "stone": 2},
		XPValue:       15,
	})
	RegisterStructure(&StructureDef{
		Type:          STRUCT_CHEST,
		Name:          "Storage Chest",
		Description:   "Stores up to 27 items",
		Health:        40,
		Defense:       1,
		IsSolid:       false,
		IsInteractive: true,
		Cost:          map[string]int{"wood": 8},
		XPValue:       8,
	})
	RegisterStructure(&StructureDef{
		Type:          STRUCT_BED,
		Name:          "Bed",
		Description:   "Sets spawn point and skips night",
		Health:        30,
		Defense:       0,
		IsSolid:       false,
		IsInteractive: true,
		Cost:          map[string]int{"wood": 3, "wool": 3},
		XPValue:       12,
	})
	RegisterStructure(&StructureDef{
		Type:          STRUCT_LAMP,
		Name:          "Lantern",
		Description:   "Provides light",
		Health:        20,
		Defense:       0,
		IsSolid:       false,
		IsInteractive: true,
		Cost:          map[string]int{"iron": 2, "coal": 1},
		XPValue:       5,
	})
}
