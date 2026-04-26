package effects

import (
	"time"
)

// EffectType represents different types of status effects
type EffectType int

const (
	EFFECT_NONE EffectType = iota
	EFFECT_POISON
	EFFECT_REGENERATION
	EFFECT_SPEED
	EFFECT_SLOWNESS
	EFFECT_STRENGTH
	EFFECT_WEAKNESS
	EFFECT_HUNGER
	EFFECT_THIRST
	EFFECT_BLEEDING
	EFFECT_NIGHT_VISION
	EFFECT_MINING_BOOST
	EFFECT_DEFENSE
)

// EffectModifier represents how an effect modifies stats
type EffectModifier struct {
	HealthRegen      float64 // HP per second
	HungerDrain      float64 // Hunger per second
	ThirstDrain      float64 // Thirst per second
	StaminaRegen     float64 // Stamina per second
	MovementSpeed    float64 // Multiplier (1.0 = normal)
	MiningSpeed      float64 // Multiplier
	DamageMultiplier float64 // For combat
	DefenseBonus     float64 // Flat defense increase
	VisionRange      float64 // Multiplier for view distance
}

// EffectDefinition defines a status effect type
type EffectDefinition struct {
	Type         EffectType
	Name         string
	Description  string
	Icon         string
	IsNegative   bool
	IsStackable  bool
	MaxStacks    int
	BaseModifier EffectModifier
}

// ActiveEffect represents an active instance of an effect
type ActiveEffect struct {
	Definition    *EffectDefinition
	StartTime     time.Time
	Duration      time.Duration
	Stacks        int
	RemainingTime time.Duration
}

// IsExpired returns true if the effect has expired
func (ae *ActiveEffect) IsExpired() bool {
	return time.Since(ae.StartTime) >= ae.Duration
}

// GetRemainingDuration returns the remaining time
func (ae *ActiveEffect) GetRemainingDuration() time.Duration {
	elapsed := time.Since(ae.StartTime)
	if elapsed >= ae.Duration {
		return 0
	}
	return ae.Duration - elapsed
}

// GetProgress returns progress from 0.0 (start) to 1.0 (expired)
func (ae *ActiveEffect) GetProgress() float64 {
	elapsed := time.Since(ae.StartTime)
	if elapsed >= ae.Duration {
		return 1.0
	}
	return float64(elapsed) / float64(ae.Duration)
}

// EffectRegistry holds all effect definitions
var EffectRegistry = make(map[EffectType]*EffectDefinition)

// RegisterEffect registers an effect definition
func RegisterEffect(def *EffectDefinition) {
	EffectRegistry[def.Type] = def
}

// GetEffectDefinition retrieves an effect definition by type
func GetEffectDefinition(effectType EffectType) *EffectDefinition {
	return EffectRegistry[effectType]
}

func init() {
	// Register default effects
	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_POISON,
		Name:        "Poison",
		Description: "Losing health over time",
		Icon:        "poison",
		IsNegative:  true,
		IsStackable: true,
		MaxStacks:   3,
		BaseModifier: EffectModifier{
			HealthRegen: -2.0,
			HungerDrain: 0.5,
		},
	})

	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_REGENERATION,
		Name:        "Regeneration",
		Description: "Recovering health over time",
		Icon:        "regen",
		IsNegative:  false,
		IsStackable: false,
		MaxStacks:   1,
		BaseModifier: EffectModifier{
			HealthRegen: 1.0,
		},
	})

	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_SPEED,
		Name:        "Speed",
		Description: "Moving faster",
		Icon:        "speed",
		IsNegative:  false,
		IsStackable: false,
		MaxStacks:   1,
		BaseModifier: EffectModifier{
			MovementSpeed: 1.3,
		},
	})

	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_SLOWNESS,
		Name:        "Slowness",
		Description: "Moving slower",
		Icon:        "slow",
		IsNegative:  true,
		IsStackable: false,
		MaxStacks:   1,
		BaseModifier: EffectModifier{
			MovementSpeed: 0.7,
		},
	})

	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_STRENGTH,
		Name:        "Strength",
		Description: "Dealing more damage",
		Icon:        "strength",
		IsNegative:  false,
		IsStackable: false,
		MaxStacks:   1,
		BaseModifier: EffectModifier{
			DamageMultiplier: 1.5,
			MiningSpeed:      1.3,
		},
	})

	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_WEAKNESS,
		Name:        "Weakness",
		Description: "Dealing less damage",
		Icon:        "weakness",
		IsNegative:  true,
		IsStackable: false,
		MaxStacks:   1,
		BaseModifier: EffectModifier{
			DamageMultiplier: 0.5,
			MiningSpeed:      0.7,
		},
	})

	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_HUNGER,
		Name:        "Hunger",
		Description: "Getting hungrier faster",
		Icon:        "hunger",
		IsNegative:  true,
		IsStackable: true,
		MaxStacks:   3,
		BaseModifier: EffectModifier{
			HungerDrain: 2.0,
		},
	})

	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_THIRST,
		Name:        "Thirst",
		Description: "Getting thirstier faster",
		Icon:        "thirst",
		IsNegative:  true,
		IsStackable: true,
		MaxStacks:   3,
		BaseModifier: EffectModifier{
			ThirstDrain: 2.0,
		},
	})

	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_BLEEDING,
		Name:        "Bleeding",
		Description: "Losing health rapidly",
		Icon:        "bleed",
		IsNegative:  true,
		IsStackable: true,
		MaxStacks:   3,
		BaseModifier: EffectModifier{
			HealthRegen:  -5.0,
			StaminaRegen: -2.0,
		},
	})

	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_NIGHT_VISION,
		Name:        "Night Vision",
		Description: "See better in darkness",
		Icon:        "night",
		IsNegative:  false,
		IsStackable: false,
		MaxStacks:   1,
		BaseModifier: EffectModifier{
			VisionRange: 1.5,
		},
	})

	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_MINING_BOOST,
		Name:        "Mining Boost",
		Description: "Mining blocks faster",
		Icon:        "mining",
		IsNegative:  false,
		IsStackable: false,
		MaxStacks:   1,
		BaseModifier: EffectModifier{
			MiningSpeed: 2.0,
		},
	})

	RegisterEffect(&EffectDefinition{
		Type:        EFFECT_DEFENSE,
		Name:        "Defense",
		Description: "Taking less damage",
		Icon:        "defense",
		IsNegative:  false,
		IsStackable: false,
		MaxStacks:   1,
		BaseModifier: EffectModifier{
			DefenseBonus: 5.0,
		},
	})
}
