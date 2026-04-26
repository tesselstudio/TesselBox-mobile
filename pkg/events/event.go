package events

import (
	"math/rand"
	"time"
)

// EventType represents different types of random events
type EventType int

const (
	EVENT_NONE EventType = iota
	EVENT_STORM
	EVENT_EARTHQUAKE
	EVENT_METEOR_SHOWER
	EVENT_RESOURCE_BOOM
	EVENT_FOG
	EVENT_BLOOD_MOON
	EVENT_SOLAR_ECLIPSE
	EVENT_WIND_STORM
	EVENT_ACID_RAIN
	EVENT_LOOT_RAIN
)

// EventSeverity represents how dangerous/intense an event is
type EventSeverity int

const (
	SEVERITY_MILD EventSeverity = iota
	SEVERITY_MODERATE
	SEVERITY_SEVERE
	SEVERITY_EXTREME
)

func (es EventSeverity) String() string {
	switch es {
	case SEVERITY_MILD:
		return "Mild"
	case SEVERITY_MODERATE:
		return "Moderate"
	case SEVERITY_SEVERE:
		return "Severe"
	case SEVERITY_EXTREME:
		return "EXTREME"
	default:
		return "Unknown"
	}
}

// EventPhase represents the current phase of an event
type EventPhase int

const (
	PHASE_WARNING  EventPhase = iota // Event is approaching
	PHASE_ACTIVE                     // Event is happening
	PHASE_ENDING                     // Event is winding down
	PHASE_FINISHED                   // Event is done
)

// EventEffect represents what an event does
type EventEffect struct {
	Type          string
	Magnitude     float64
	Duration      time.Duration
	AffectedStats []string
}

// EventDefinition defines a random event type
type EventDefinition struct {
	Type           EventType
	Name           string
	Description    string
	WarningMessage string
	ActiveMessage  string
	EndMessage     string
	Severity       EventSeverity
	MinDuration    time.Duration
	MaxDuration    time.Duration
	WarningTime    time.Duration // Time before event starts
	CooldownTime   time.Duration // Minimum time before this event can happen again
	Effects        []EventEffect
	Requirements   EventRequirements
	Weight         int // Chance weight (higher = more common)
}

// EventRequirements defines what conditions must be met for an event
type EventRequirements struct {
	MinDay         int    // Minimum day count
	MinPlayTime    int    // Minimum minutes played
	RequireSurface bool   // Must be on surface layer
	TimeOfDay      string // "day", "night", or "any"
	Weather        string // Required weather condition
}

// ActiveEvent represents an active event instance
type ActiveEvent struct {
	Definition  *EventDefinition
	Phase       EventPhase
	StartTime   time.Time
	EndTime     time.Time
	WarningTime time.Time
	Intensity   float64 // 0-1 scale
	CustomData  map[string]interface{}
}

// IsActive returns true if event is currently active
func (ae *ActiveEvent) IsActive() bool {
	return ae.Phase == PHASE_ACTIVE || ae.Phase == PHASE_WARNING
}

// GetRemainingTime returns time until event ends
func (ae *ActiveEvent) GetRemainingTime() time.Duration {
	if ae.Phase == PHASE_FINISHED {
		return 0
	}
	remaining := time.Until(ae.EndTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetProgress returns progress from 0 (start) to 1 (end)
func (ae *ActiveEvent) GetProgress() float64 {
	totalDuration := ae.EndTime.Sub(ae.StartTime)
	elapsed := time.Since(ae.StartTime)
	if elapsed >= totalDuration {
		return 1.0
	}
	return float64(elapsed) / float64(totalDuration)
}

// GetTimeUntilStart returns time until event starts (during warning phase)
func (ae *ActiveEvent) GetTimeUntilStart() time.Duration {
	if ae.Phase != PHASE_WARNING {
		return 0
	}
	remaining := time.Until(ae.StartTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// EventRegistry holds all event definitions
var EventRegistry = make(map[EventType]*EventDefinition)

// RegisterEvent registers an event definition
func RegisterEvent(def *EventDefinition) {
	EventRegistry[def.Type] = def
}

// GetEventDefinition retrieves an event definition
func GetEventDefinition(eventType EventType) *EventDefinition {
	return EventRegistry[eventType]
}

// GetRandomEvent returns a random event based on weights
func GetRandomEvent(day int, playTime int, isSurface bool, timeOfDay string) *EventDefinition {
	var validEvents []*EventDefinition
	var totalWeight int

	for _, def := range EventRegistry {
		if def.MeetsRequirements(day, playTime, isSurface, timeOfDay) {
			validEvents = append(validEvents, def)
			totalWeight += def.Weight
		}
	}

	if len(validEvents) == 0 {
		return nil
	}

	// Weighted random selection
	randomValue := rand.Intn(totalWeight)
	currentWeight := 0

	for _, def := range validEvents {
		currentWeight += def.Weight
		if randomValue < currentWeight {
			return def
		}
	}

	return validEvents[0]
}

// MeetsRequirements checks if requirements are met
func (ed *EventDefinition) MeetsRequirements(day, playTime int, isSurface bool, timeOfDay string) bool {
	req := ed.Requirements

	if day < req.MinDay {
		return false
	}
	if playTime < req.MinPlayTime {
		return false
	}
	if req.RequireSurface && !isSurface {
		return false
	}
	if req.TimeOfDay != "" && req.TimeOfDay != "any" && req.TimeOfDay != timeOfDay {
		return false
	}

	return true
}

func init() {
	// Register Storm event
	RegisterEvent(&EventDefinition{
		Type:           EVENT_STORM,
		Name:           "Thunder Storm",
		Description:    "Heavy rain and lightning",
		WarningMessage: "Dark clouds gather...",
		ActiveMessage:  "A storm is raging!",
		EndMessage:     "The storm subsides.",
		Severity:       SEVERITY_MODERATE,
		MinDuration:    2 * time.Minute,
		MaxDuration:    5 * time.Minute,
		WarningTime:    30 * time.Second,
		CooldownTime:   10 * time.Minute,
		Weight:         30,
		Requirements: EventRequirements{
			TimeOfDay: "any",
		},
		Effects: []EventEffect{
			{Type: "mining_slow", Magnitude: 0.5, Duration: 2 * time.Minute, AffectedStats: []string{"mining_speed"}},
			{Type: "visibility", Magnitude: 0.7, Duration: 5 * time.Minute, AffectedStats: []string{"vision"}},
		},
	})

	// Register Earthquake event
	RegisterEvent(&EventDefinition{
		Type:           EVENT_EARTHQUAKE,
		Name:           "Earthquake",
		Description:    "The ground shakes violently",
		WarningMessage: "The ground begins to tremble...",
		ActiveMessage:  "EARTHQUAKE! Hold on!",
		EndMessage:     "The tremors stop.",
		Severity:       SEVERITY_SEVERE,
		MinDuration:    30 * time.Second,
		MaxDuration:    2 * time.Minute,
		WarningTime:    5 * time.Second,
		CooldownTime:   15 * time.Minute,
		Weight:         10,
		Requirements: EventRequirements{
			MinDay: 3,
		},
		Effects: []EventEffect{
			{Type: "movement_slow", Magnitude: 0.3, Duration: 2 * time.Minute, AffectedStats: []string{"movement_speed"}},
			{Type: "mining_stop", Magnitude: 1.0, Duration: 2 * time.Minute, AffectedStats: []string{"mining_enabled"}},
		},
	})

	// Register Meteor Shower event
	RegisterEvent(&EventDefinition{
		Type:           EVENT_METEOR_SHOWER,
		Name:           "Meteor Shower",
		Description:    "Meteors fall from the sky",
		WarningMessage: "Strange lights appear in the sky...",
		ActiveMessage:  "Meteor shower incoming!",
		EndMessage:     "The meteor shower ends.",
		Severity:       SEVERITY_EXTREME,
		MinDuration:    1 * time.Minute,
		MaxDuration:    3 * time.Minute,
		WarningTime:    45 * time.Second,
		CooldownTime:   30 * time.Minute,
		Weight:         5,
		Requirements: EventRequirements{
			MinDay:         5,
			RequireSurface: true,
			TimeOfDay:      "night",
		},
		Effects: []EventEffect{
			{Type: "meteor_strike", Magnitude: 1.0, Duration: 3 * time.Minute, AffectedStats: []string{}},
		},
	})

	// Register Resource Boom event
	RegisterEvent(&EventDefinition{
		Type:           EVENT_RESOURCE_BOOM,
		Name:           "Resource Boom",
		Description:    "Resources are more abundant",
		WarningMessage: "You sense something special in the air...",
		ActiveMessage:  "RESOURCE BOOM! Extra drops!",
		EndMessage:     "The resource boom fades.",
		Severity:       SEVERITY_MILD,
		MinDuration:    3 * time.Minute,
		MaxDuration:    8 * time.Minute,
		WarningTime:    15 * time.Second,
		CooldownTime:   20 * time.Minute,
		Weight:         15,
		Requirements: EventRequirements{
			MinDay: 2,
		},
		Effects: []EventEffect{
			{Type: "double_drops", Magnitude: 2.0, Duration: 8 * time.Minute, AffectedStats: []string{"drop_rate"}},
		},
	})

	// Register Fog event
	RegisterEvent(&EventDefinition{
		Type:           EVENT_FOG,
		Name:           "Thick Fog",
		Description:    "Visibility is severely reduced",
		WarningMessage: "A thick fog rolls in...",
		ActiveMessage:  "Thick fog surrounds you!",
		EndMessage:     "The fog clears.",
		Severity:       SEVERITY_MILD,
		MinDuration:    4 * time.Minute,
		MaxDuration:    10 * time.Minute,
		WarningTime:    20 * time.Second,
		CooldownTime:   8 * time.Minute,
		Weight:         25,
		Requirements: EventRequirements{
			TimeOfDay: "night",
		},
		Effects: []EventEffect{
			{Type: "visibility", Magnitude: 0.4, Duration: 10 * time.Minute, AffectedStats: []string{"vision"}},
		},
	})

	// Register Blood Moon event
	RegisterEvent(&EventDefinition{
		Type:           EVENT_BLOOD_MOON,
		Name:           "Blood Moon",
		Description:    "Enemies are stronger and more aggressive",
		WarningMessage: "The moon turns blood red...",
		ActiveMessage:  "BLOOD MOON RISING! Danger approaches!",
		EndMessage:     "The blood moon sets.",
		Severity:       SEVERITY_SEVERE,
		MinDuration:    5 * time.Minute,
		MaxDuration:    10 * time.Minute,
		WarningTime:    1 * time.Minute,
		CooldownTime:   45 * time.Minute,
		Weight:         8,
		Requirements: EventRequirements{
			MinDay:    7,
			TimeOfDay: "night",
		},
		Effects: []EventEffect{
			{Type: "enemy_strength", Magnitude: 1.5, Duration: 10 * time.Minute, AffectedStats: []string{"enemy_damage", "enemy_speed"}},
		},
	})

	// Register Loot Rain event (positive)
	RegisterEvent(&EventDefinition{
		Type:           EVENT_LOOT_RAIN,
		Name:           "Loot Rain",
		Description:    "Valuable items rain from the sky",
		WarningMessage: "You hear the sound of falling treasures...",
		ActiveMessage:  "LOOT RAIN! Catch what you can!",
		EndMessage:     "The loot rain stops.",
		Severity:       SEVERITY_MILD,
		MinDuration:    1 * time.Minute,
		MaxDuration:    2 * time.Minute,
		WarningTime:    10 * time.Second,
		CooldownTime:   25 * time.Minute,
		Weight:         5,
		Requirements: EventRequirements{
			MinDay:         4,
			RequireSurface: true,
		},
		Effects: []EventEffect{
			{Type: "item_spawn", Magnitude: 1.0, Duration: 2 * time.Minute, AffectedStats: []string{}},
		},
	})
}
