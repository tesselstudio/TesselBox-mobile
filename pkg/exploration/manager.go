package exploration

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ExplorationManager handles POI discovery and management
type ExplorationManager struct {
	pois            []*PointOfInterest
	mu              sync.RWMutex
	discoveredCount int
	detectionRange  float64
	onDiscover      func(poi *PointOfInterest)
	onExplore       func(poi *PointOfInterest, xp int, rewards []Reward)
}

// NewExplorationManager creates a new exploration manager
func NewExplorationManager() *ExplorationManager {
	return &ExplorationManager{
		pois:           make([]*PointOfInterest, 0),
		detectionRange: 100.0,
	}
}

// SetCallbacks sets event callbacks
func (em *ExplorationManager) SetCallbacks(
	onDiscover func(poi *PointOfInterest),
	onExplore func(poi *PointOfInterest, xp int, rewards []Reward),
) {
	em.onDiscover = onDiscover
	em.onExplore = onExplore
}

// GeneratePOIs creates random POIs in a world area
func (em *ExplorationManager) GeneratePOIs(worldWidth, worldHeight float64, layer int, count int) {
	em.mu.Lock()
	defer em.mu.Unlock()

	poiTypes := []POIType{
		POI_RUINS, POI_CAVE, POI_ANCIENT_STRUCTURE,
		POI_LOOT_CACHE, POI_ABANDONED_MINE, POI_LANDMARK,
	}

	for i := 0; i < count; i++ {
		poiType := poiTypes[rand.Intn(len(poiTypes))]
		template := POIRegistry[poiType]
		if template == nil {
			continue
		}

		poi := &PointOfInterest{
			ID:          generatePOIID(i),
			Type:        poiType,
			Name:        template.Name,
			Description: template.Description,
			Rarity:      template.Rarity,
			X:           rand.Float64() * worldWidth,
			Y:           rand.Float64() * worldHeight,
			Layer:       layer,
			Range:       template.Range,
			XPValue:     template.XPValue,
			Rewards:     template.Rewards,
			Discovered:  false,
			Explored:    false,
		}
		em.pois = append(em.pois, poi)
	}
}

// Update checks for POI discoveries
func (em *ExplorationManager) Update(playerX, playerY float64, playerLayer int) {
	em.mu.Lock()
	defer em.mu.Unlock()

	for _, poi := range em.pois {
		if !poi.Discovered && poi.CanDiscover(playerX, playerY, playerLayer, em.detectionRange) {
			poi.Discover()
			em.discoveredCount++

			if em.onDiscover != nil {
				em.onDiscover(poi)
			}
		}
	}
}

// ExplorePOI explores a discovered POI
func (em *ExplorationManager) ExplorePOI(poiID string) (int, []Reward) {
	em.mu.Lock()
	defer em.mu.Unlock()

	for _, poi := range em.pois {
		if poi.ID == poiID && poi.Discovered {
			xp, rewards := poi.Explore()
			if xp > 0 && em.onExplore != nil {
				em.onExplore(poi, xp, rewards)
			}
			return xp, rewards
		}
	}
	return 0, nil
}

// GetNearbyPOIs returns POIs near a location
func (em *ExplorationManager) GetNearbyPOIs(x, y float64, layer int, radius float64) []*PointOfInterest {
	em.mu.RLock()
	defer em.mu.RUnlock()

	radiusSquared := radius * radius
	var nearby []*PointOfInterest

	for _, poi := range em.pois {
		if poi.Layer != layer {
			continue
		}
		distance := poi.GetDistance(x, y)
		if distance <= radiusSquared {
			nearby = append(nearby, poi)
		}
	}
	return nearby
}

// GetDiscoveredPOIs returns all discovered POIs
func (em *ExplorationManager) GetDiscoveredPOIs() []*PointOfInterest {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var discovered []*PointOfInterest
	for _, poi := range em.pois {
		if poi.Discovered {
			discovered = append(discovered, poi)
		}
	}
	return discovered
}

// GetStats returns exploration statistics
func (em *ExplorationManager) GetStats() (discovered, total int) {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.discoveredCount, len(em.pois)
}

// SetDetectionRange modifies detection range
func (em *ExplorationManager) SetDetectionRange(rangeVal float64) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.detectionRange = rangeVal
}

// RemovePOI removes a POI by ID
func (em *ExplorationManager) RemovePOI(poiID string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	for i, poi := range em.pois {
		if poi.ID == poiID {
			em.pois = append(em.pois[:i], em.pois[i+1:]...)
			if poi.Discovered {
				em.discoveredCount--
			}
			return
		}
	}
}

// Clear clears all POIs
func (em *ExplorationManager) Clear() {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.pois = make([]*PointOfInterest, 0)
	em.discoveredCount = 0
}

// generatePOIID generates a unique ID
func generatePOIID(index int) string {
	return fmt.Sprintf("poi_%d_%d", time.Now().Unix(), index)
}
