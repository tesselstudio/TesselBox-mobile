package effects

import (
	"sync"
	"time"
)

// EffectManager manages all active status effects on the player
type EffectManager struct {
	activeEffects     map[EffectType]*ActiveEffect
	mu                sync.RWMutex
	onEffectApplied   func(effect *ActiveEffect)
	onEffectExpired   func(effect *ActiveEffect)
	onEffectRefreshed func(effect *ActiveEffect)
}

// NewEffectManager creates a new effect manager
func NewEffectManager() *EffectManager {
	return &EffectManager{
		activeEffects: make(map[EffectType]*ActiveEffect),
	}
}

// SetCallbacks sets event callbacks
func (em *EffectManager) SetCallbacks(
	onApplied func(effect *ActiveEffect),
	onExpired func(effect *ActiveEffect),
	onRefreshed func(effect *ActiveEffect),
) {
	em.onEffectApplied = onApplied
	em.onEffectExpired = onExpired
	em.onEffectRefreshed = onRefreshed
}

// ApplyEffect applies a new effect or refreshes existing
func (em *EffectManager) ApplyEffect(effectType EffectType, duration time.Duration, stacks int) bool {
	def := GetEffectDefinition(effectType)
	if def == nil {
		return false
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	existing, hasExisting := em.activeEffects[effectType]

	if hasExisting {
		// Refresh existing effect
		if def.IsStackable && existing.Stacks < def.MaxStacks {
			existing.Stacks = minInt(existing.Stacks+stacks, def.MaxStacks)
		}
		// Extend duration
		elapsed := time.Since(existing.StartTime)
		existing.Duration = maxDuration(existing.Duration, elapsed+duration)
		existing.StartTime = time.Now()

		if em.onEffectRefreshed != nil {
			em.onEffectRefreshed(existing)
		}
	} else {
		// Apply new effect
		effect := &ActiveEffect{
			Definition: def,
			StartTime:  time.Now(),
			Duration:   duration,
			Stacks:     minInt(stacks, def.MaxStacks),
		}
		em.activeEffects[effectType] = effect

		if em.onEffectApplied != nil {
			em.onEffectApplied(effect)
		}
	}

	return true
}

// RemoveEffect removes a specific effect
func (em *EffectManager) RemoveEffect(effectType EffectType) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	effect, exists := em.activeEffects[effectType]
	if !exists {
		return false
	}

	delete(em.activeEffects, effectType)

	if em.onEffectExpired != nil {
		em.onEffectExpired(effect)
	}

	return true
}

// RemoveAllEffects removes all active effects
func (em *EffectManager) RemoveAllEffects() {
	em.mu.Lock()
	defer em.mu.Unlock()

	for _, effect := range em.activeEffects {
		if em.onEffectExpired != nil {
			em.onEffectExpired(effect)
		}
	}

	em.activeEffects = make(map[EffectType]*ActiveEffect)
}

// RemoveNegativeEffects removes all negative effects
func (em *EffectManager) RemoveNegativeEffects() {
	em.mu.Lock()
	defer em.mu.Unlock()

	for effectType, effect := range em.activeEffects {
		if effect.Definition.IsNegative {
			delete(em.activeEffects, effectType)
			if em.onEffectExpired != nil {
				em.onEffectExpired(effect)
			}
		}
	}
}

// GetActiveEffects returns all active effects
func (em *EffectManager) GetActiveEffects() []*ActiveEffect {
	em.mu.RLock()
	defer em.mu.RUnlock()

	effects := make([]*ActiveEffect, 0, len(em.activeEffects))
	for _, effect := range em.activeEffects {
		effects = append(effects, effect)
	}
	return effects
}

// HasEffect checks if a specific effect is active
func (em *EffectManager) HasEffect(effectType EffectType) bool {
	em.mu.RLock()
	defer em.mu.RUnlock()

	_, exists := em.activeEffects[effectType]
	return exists
}

// GetEffect returns an active effect if it exists
func (em *EffectManager) GetEffect(effectType EffectType) *ActiveEffect {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return em.activeEffects[effectType]
}

// GetCombinedModifiers calculates total modifiers from all effects
func (em *EffectManager) GetCombinedModifiers() EffectModifier {
	em.mu.RLock()
	defer em.mu.RUnlock()

	combined := EffectModifier{
		HealthRegen:      0,
		HungerDrain:      0,
		ThirstDrain:      0,
		StaminaRegen:     0,
		MovementSpeed:    1.0,
		MiningSpeed:      1.0,
		DamageMultiplier: 1.0,
		DefenseBonus:     0,
		VisionRange:      1.0,
	}

	for _, effect := range em.activeEffects {
		mod := effect.Definition.BaseModifier
		stacks := float64(effect.Stacks)
		if stacks == 0 {
			stacks = 1
		}

		// Additive modifiers
		combined.HealthRegen += mod.HealthRegen * stacks
		combined.HungerDrain += mod.HungerDrain * stacks
		combined.ThirstDrain += mod.ThirstDrain * stacks
		combined.StaminaRegen += mod.StaminaRegen * stacks
		combined.DefenseBonus += mod.DefenseBonus * stacks

		// Multiplicative modifiers (multiply the modifiers)
		if mod.MovementSpeed != 0 {
			combined.MovementSpeed *= 1 + (mod.MovementSpeed-1)*stacks
		}
		if mod.MiningSpeed != 0 {
			combined.MiningSpeed *= 1 + (mod.MiningSpeed-1)*stacks
		}
		if mod.DamageMultiplier != 0 {
			combined.DamageMultiplier *= 1 + (mod.DamageMultiplier-1)*stacks
		}
		if mod.VisionRange != 0 {
			combined.VisionRange *= 1 + (mod.VisionRange-1)*stacks
		}
	}

	// Clamp multipliers to reasonable bounds
	combined.MovementSpeed = clampFloat64(combined.MovementSpeed, 0.1, 3.0)
	combined.MiningSpeed = clampFloat64(combined.MiningSpeed, 0.1, 5.0)
	combined.DamageMultiplier = clampFloat64(combined.DamageMultiplier, 0.1, 5.0)
	combined.VisionRange = clampFloat64(combined.VisionRange, 0.5, 3.0)

	return combined
}

// Update processes effect expiration
func (em *EffectManager) Update(deltaTime float64) {
	em.mu.Lock()
	defer em.mu.Unlock()

	now := time.Now()
	for effectType, effect := range em.activeEffects {
		elapsed := now.Sub(effect.StartTime)
		if elapsed >= effect.Duration {
			delete(em.activeEffects, effectType)
			if em.onEffectExpired != nil {
				go em.onEffectExpired(effect)
			}
		}
	}
}

// GetActiveEffectCount returns the number of active effects
func (em *EffectManager) GetActiveEffectCount() int {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return len(em.activeEffects)
}

// HasNegativeEffects returns true if any negative effect is active
func (em *EffectManager) HasNegativeEffects() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()

	for _, effect := range em.activeEffects {
		if effect.Definition.IsNegative {
			return true
		}
	}
	return false
}

// HasPositiveEffects returns true if any positive effect is active
func (em *EffectManager) HasPositiveEffects() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()

	for _, effect := range em.activeEffects {
		if !effect.Definition.IsNegative {
			return true
		}
	}
	return false
}

// Helper functions
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func clampFloat64(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
