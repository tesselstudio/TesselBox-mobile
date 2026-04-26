package farming

import (
	"fmt"
	"sync"
	"time"
)

// FarmingManager handles crop management
type FarmingManager struct {
	crops        []*CropInstance
	mu           sync.RWMutex
	onCropReady  func(crop *CropInstance)
	onCropDie    func(crop *CropInstance)
	onHarvest    func(crop *CropInstance, yields []CropYield)
}

// NewFarmingManager creates a new farming manager
func NewFarmingManager() *FarmingManager {
	return &FarmingManager{
		crops: make([]*CropInstance, 0),
	}
}

// SetCallbacks sets event callbacks
func (fm *FarmingManager) SetCallbacks(
	onCropReady func(crop *CropInstance),
	onCropDie func(crop *CropInstance),
	onHarvest func(crop *CropInstance, yields []CropYield),
) {
	fm.onCropReady = onCropReady
	fm.onCropDie = onCropDie
	fm.onHarvest = onHarvest
}

// PlantCrop plants a new crop
func (fm *FarmingManager) PlantCrop(cropType CropType, x, y float64, layer int) *CropInstance {
	def := CropRegistry[cropType]
	if def == nil {
		return nil
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	crop := &CropInstance{
		ID:         generateCropID(),
		Definition: def,
		X:          x,
		Y:          y,
		Layer:      layer,
		PlantedAt:  time.Now(),
		Stage:      STAGE_SEED,
		WaterLevel: 0,
		Health:     100,
		IsDead:     false,
	}

	fm.crops = append(fm.crops, crop)
	return crop
}

// Update updates all crops
func (fm *FarmingManager) Update(deltaTime float64, dayTime float64) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for _, crop := range fm.crops {
		oldStage := crop.Stage
		crop.UpdateGrowth(deltaTime, dayTime)
		
		// Check for stage changes
		if oldStage != crop.Stage && crop.Stage == STAGE_HARVESTABLE {
			if fm.onCropReady != nil {
				fm.onCropReady(crop)
			}
		}
		
		// Check for death
		if crop.IsDead && fm.onCropDie != nil {
			fm.onCropDie(crop)
		}
	}
}

// WaterCrop waters a specific crop by ID
func (fm *FarmingManager) WaterCrop(cropID string, amount int) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for _, crop := range fm.crops {
		if crop.ID == cropID && !crop.IsDead {
			crop.Water(amount)
			return true
		}
	}
	return false
}

// WaterCropsInRange waters all crops within range of a point
func (fm *FarmingManager) WaterCropsInRange(x, y float64, layer int, radius float64, amount int) int {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	watered := 0
	radiusSquared := radius * radius

	for _, crop := range fm.crops {
		if crop.Layer != layer || crop.IsDead {
			continue
		}
		dx := crop.X - x
		dy := crop.Y - y
		if dx*dx+dy*dy <= radiusSquared {
			crop.Water(amount)
			watered++
		}
	}
	return watered
}

// HarvestCrop harvests a specific crop
func (fm *FarmingManager) HarvestCrop(cropID string) []CropYield {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for _, crop := range fm.crops {
		if crop.ID == cropID && crop.CanHarvest() {
			yields := crop.Harvest()
			if fm.onHarvest != nil {
				fm.onHarvest(crop, yields)
			}
			return yields
		}
	}
	return nil
}

// GetCropsAt returns crops at a specific location
func (fm *FarmingManager) GetCropsAt(x, y float64, layer int, tolerance float64) []*CropInstance {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var found []*CropInstance
	tolSquared := tolerance * tolerance

	for _, crop := range fm.crops {
		if crop.Layer != layer {
			continue
		}
		dx := crop.X - x
		dy := crop.Y - y
		if dx*dx+dy*dy <= tolSquared {
			found = append(found, crop)
		}
	}
	return found
}

// GetCropsInLayer returns all crops in a layer
func (fm *FarmingManager) GetCropsInLayer(layer int) []*CropInstance {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var crops []*CropInstance
	for _, crop := range fm.crops {
		if crop.Layer == layer {
			crops = append(crops, crop)
		}
	}
	return crops
}

// GetReadyCrops returns harvestable crops
func (fm *FarmingManager) GetReadyCrops(layer int) []*CropInstance {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var ready []*CropInstance
	for _, crop := range fm.crops {
		if crop.Layer == layer && crop.CanHarvest() {
			ready = append(ready, crop)
		}
	}
	return ready
}

// RemoveCrop removes a crop
func (fm *FarmingManager) RemoveCrop(cropID string) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for i, crop := range fm.crops {
		if crop.ID == cropID {
			fm.crops = append(fm.crops[:i], fm.crops[i+1:]...)
			return true
		}
	}
	return false
}

// ClearLayer removes all crops from a layer
func (fm *FarmingManager) ClearLayer(layer int) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	var remaining []*CropInstance
	for _, crop := range fm.crops {
		if crop.Layer != layer {
			remaining = append(remaining, crop)
		}
	}
	fm.crops = remaining
}

// GetStats returns farming statistics
func (fm *FarmingManager) GetStats(layer int) (total, ready, dead int) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for _, crop := range fm.crops {
		if layer >= 0 && crop.Layer != layer {
			continue
		}
		total++
		if crop.CanHarvest() {
			ready++
		}
		if crop.IsDead {
			dead++
		}
	}
	return
}

// generateCropID generates unique ID
func generateCropID() string {
	return fmt.Sprintf("crop_%d", time.Now().UnixNano())
}
