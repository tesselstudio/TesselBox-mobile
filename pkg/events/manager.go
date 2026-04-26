package events

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

// EventManager handles random event scheduling and execution
type EventManager struct {
	activeEvent    *ActiveEvent
	mu             sync.RWMutex
	lastEventTime  map[EventType]time.Time
	day            int
	playTime       int // minutes
	isSurface      bool
	timeOfDay      string // "day" or "night"
	checkInterval  time.Duration
	lastCheck      time.Time
	onEventStart   func(event *ActiveEvent)
	onEventWarning func(event *ActiveEvent)
	onEventEnd     func(event *ActiveEvent)
	baseChance     float64 // Base chance per check (0-1)
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	return &EventManager{
		lastEventTime: make(map[EventType]time.Time),
		checkInterval: 30 * time.Second,
		lastCheck:     time.Now(),
		baseChance:    0.1, // 10% base chance per check
		timeOfDay:     "day",
		isSurface:     true,
	}
}

// SetCallbacks sets event callbacks
func (em *EventManager) SetCallbacks(
	onEventStart func(event *ActiveEvent),
	onEventWarning func(event *ActiveEvent),
	onEventEnd func(event *ActiveEvent),
) {
	em.onEventStart = onEventStart
	em.onEventWarning = onEventWarning
	em.onEventEnd = onEventEnd
}

// Update updates event state and checks for new events
func (em *EventManager) Update(deltaTime float64, day int, playTime int, isSurface bool, timeOfDay string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Update state
	em.day = day
	em.playTime = playTime
	em.isSurface = isSurface
	em.timeOfDay = timeOfDay

	// Update active event
	if em.activeEvent != nil {
		em.updateActiveEvent()
	}

	// Check for new events
	if em.activeEvent == nil && time.Since(em.lastCheck) >= em.checkInterval {
		em.lastCheck = time.Now()
		em.checkForNewEvent()
	}
}

// updateActiveEvent progresses the active event through its phases
func (em *EventManager) updateActiveEvent() {
	event := em.activeEvent
	now := time.Now()

	switch event.Phase {
	case PHASE_WARNING:
		if now.After(event.StartTime) {
			// Transition to active
			event.Phase = PHASE_ACTIVE
			log.Printf("Event started: %s", event.Definition.Name)
			if em.onEventStart != nil {
				em.onEventStart(event)
			}
		}

	case PHASE_ACTIVE:
		if now.After(event.EndTime) {
			// Transition to ending
			event.Phase = PHASE_ENDING
			log.Printf("Event ending: %s", event.Definition.Name)
			// Immediately finish for now
			event.Phase = PHASE_FINISHED
			if em.onEventEnd != nil {
				em.onEventEnd(event)
			}
			em.activeEvent = nil
		}
	}
}

// checkForNewEvent attempts to trigger a random event
func (em *EventManager) checkForNewEvent() {
	// Calculate chance based on time since last event
	chance := em.baseChance

	// Increase chance if no event for a while
	timeSinceLastEvent := em.getTimeSinceLastEvent()
	if timeSinceLastEvent > 10*time.Minute {
		chance += 0.05 * float64(timeSinceLastEvent/(5*time.Minute))
	}

	// Cap chance at 30%
	if chance > 0.3 {
		chance = 0.3
	}

	// Roll for event
	if rand.Float64() >= chance {
		return // No event this time
	}

	// Select random event
	def := GetRandomEvent(em.day, em.playTime, em.isSurface, em.timeOfDay)
	if def == nil {
		return
	}

	// Check cooldown
	if lastTime, exists := em.lastEventTime[def.Type]; exists {
		if time.Since(lastTime) < def.CooldownTime {
			return // Still on cooldown
		}
	}

	// Start the event
	em.startEvent(def)
}

// startEvent initializes a new event
func (em *EventManager) startEvent(def *EventDefinition) {
	duration := def.MinDuration + time.Duration(rand.Int63n(int64(def.MaxDuration-def.MinDuration)))

	now := time.Now()
	event := &ActiveEvent{
		Definition:  def,
		Phase:       PHASE_WARNING,
		StartTime:   now.Add(def.WarningTime),
		EndTime:     now.Add(def.WarningTime + duration),
		WarningTime: now,
		Intensity:   em.calculateIntensity(def.Severity),
		CustomData:  make(map[string]interface{}),
	}

	em.activeEvent = event
	em.lastEventTime[def.Type] = now

	log.Printf("Event warning: %s - %s", def.Name, def.WarningMessage)

	if em.onEventWarning != nil {
		em.onEventWarning(event)
	}
}

// calculateIntensity calculates event intensity based on severity and randomness
func (em *EventManager) calculateIntensity(severity EventSeverity) float64 {
	baseIntensity := float64(severity) / float64(SEVERITY_EXTREME)

	// Add some randomness
	variation := (rand.Float64() - 0.5) * 0.3
	intensity := baseIntensity + variation

	// Clamp between 0.1 and 1.0
	if intensity < 0.1 {
		intensity = 0.1
	}
	if intensity > 1.0 {
		intensity = 1.0
	}

	return intensity
}

// getTimeSinceLastEvent returns time since any event occurred
func (em *EventManager) getTimeSinceLastEvent() time.Duration {
	var lastEvent time.Time

	for _, t := range em.lastEventTime {
		if t.After(lastEvent) {
			lastEvent = t
		}
	}

	if lastEvent.IsZero() {
		return 24 * time.Hour // Return large duration if no events yet
	}

	return time.Since(lastEvent)
}

// GetActiveEvent returns the currently active event
func (em *EventManager) GetActiveEvent() *ActiveEvent {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.activeEvent
}

// ForceEvent forces a specific event to start (for testing/debugging)
func (em *EventManager) ForceEvent(eventType EventType) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	def := GetEventDefinition(eventType)
	if def == nil {
		return false
	}

	em.startEvent(def)
	return true
}

// CancelActiveEvent cancels the current event
func (em *EventManager) CancelActiveEvent() {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.activeEvent != nil {
		if em.onEventEnd != nil {
			em.onEventEnd(em.activeEvent)
		}
		em.activeEvent = nil
	}
}

// GetEventHistory returns time since each event type occurred
func (em *EventManager) GetEventHistory() map[EventType]time.Duration {
	em.mu.RLock()
	defer em.mu.RUnlock()

	history := make(map[EventType]time.Duration)
	for eventType, lastTime := range em.lastEventTime {
		history[eventType] = time.Since(lastTime)
	}
	return history
}

// SetBaseChance modifies the base event chance
func (em *EventManager) SetBaseChance(chance float64) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if chance < 0 {
		chance = 0
	}
	if chance > 1 {
		chance = 1
	}
	em.baseChance = chance
}

// IsEventActive returns true if any event is currently active
func (em *EventManager) IsEventActive() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.activeEvent != nil
}

// GetActiveEffects returns current event effects
func (em *EventManager) GetActiveEffects() []EventEffect {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.activeEvent == nil || em.activeEvent.Phase != PHASE_ACTIVE {
		return nil
	}

	// Apply intensity scaling to effects
	var scaledEffects []EventEffect
	intensity := em.activeEvent.Intensity

	for _, effect := range em.activeEvent.Definition.Effects {
		scaled := effect
		scaled.Magnitude = effect.Magnitude * (0.5 + 0.5*intensity) // Scale 50%-100% of base
		scaledEffects = append(scaledEffects, scaled)
	}

	return scaledEffects
}

// GetCurrentModifiers returns stat modifiers from active event
func (em *EventManager) GetCurrentModifiers() map[string]float64 {
	effects := em.GetActiveEffects()
	if effects == nil {
		return nil
	}

	modifiers := make(map[string]float64)
	for _, effect := range effects {
		for _, stat := range effect.AffectedStats {
			modifiers[stat] = effect.Magnitude
		}
	}

	return modifiers
}
