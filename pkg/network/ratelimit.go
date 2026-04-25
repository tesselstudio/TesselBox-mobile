package network

import (
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu            sync.Mutex
	tokens        float64
	capacity      float64
	refillRate    float64 // tokens per second
	lastRefill    time.Time
	violationCount int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(capacity, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.lastRefill = now
	
	// Refill tokens
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.capacity {
		rl.tokens = rl.capacity
	}
	
	// Check if we have enough tokens
	if rl.tokens >= 1 {
		rl.tokens -= 1
		return true
	}
	
	rl.violationCount++
	return false
}

// GetViolationCount returns the number of violations
func (rl *RateLimiter) GetViolationCount() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.violationCount
}

// ResetViolations resets the violation counter
func (rl *RateLimiter) ResetViolations() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.violationCount = 0
}

// PlayerRateLimiter manages rate limiting per player
type PlayerRateLimiter struct {
	limits map[string]*RateLimiter
	mu     sync.RWMutex
	
	// Default limits
	defaultPacketsPerSecond float64
	defaultBurstSize       float64
}

// NewPlayerRateLimiter creates a new player rate limiter
func NewPlayerRateLimiter(packetsPerSecond, burstSize float64) *PlayerRateLimiter {
	return &PlayerRateLimiter{
		limits:                  make(map[string]*RateLimiter),
		defaultPacketsPerSecond: packetsPerSecond,
		defaultBurstSize:       burstSize,
	}
}

// Allow checks if a player's request is allowed
func (prl *PlayerRateLimiter) Allow(playerID string) bool {
	prl.mu.RLock()
	limiter, exists := prl.limits[playerID]
	prl.mu.RUnlock()
	
	if !exists {
		prl.mu.Lock()
		limiter = NewRateLimiter(prl.defaultBurstSize, prl.defaultPacketsPerSecond)
		prl.limits[playerID] = limiter
		prl.mu.Unlock()
	}
	
	return limiter.Allow()
}

// GetViolationCount returns violation count for a player
func (prl *PlayerRateLimiter) GetViolationCount(playerID string) int {
	prl.mu.RLock()
	defer prl.mu.RUnlock()
	
	if limiter, exists := prl.limits[playerID]; exists {
		return limiter.GetViolationCount()
	}
	return 0
}

// RemovePlayer removes a player from rate limiting
func (prl *PlayerRateLimiter) RemovePlayer(playerID string) {
	prl.mu.Lock()
	defer prl.mu.Unlock()
	delete(prl.limits, playerID)
}

// CleanupInactive removes inactive players
func (prl *PlayerRateLimiter) CleanupInactive() {
	prl.mu.Lock()
	defer prl.mu.Unlock()
	
	// In a real implementation, we'd track last activity time
	// For now, this is a placeholder
}

// SetCustomLimits sets custom rate limits for a player
func (prl *PlayerRateLimiter) SetCustomLimits(playerID string, packetsPerSecond, burstSize float64) {
	prl.mu.Lock()
	defer prl.mu.Unlock()
	
	prl.limits[playerID] = NewRateLimiter(burstSize, packetsPerSecond)
}
