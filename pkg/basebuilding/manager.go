package basebuilding

import (
	"fmt"
	"sync"
	"time"
)

// BaseManager handles player structures
type BaseManager struct {
	mu         sync.RWMutex
	structures []*StructureInstance
	onBuild    func(structure *StructureInstance)
	onDestroy  func(structure *StructureInstance)
}

// NewBaseManager creates new base manager
func NewBaseManager() *BaseManager {
	return &BaseManager{
		structures: make([]*StructureInstance, 0),
	}
}

// SetCallbacks sets event callbacks
func (bm *BaseManager) SetCallbacks(
	onBuild func(structure *StructureInstance),
	onDestroy func(structure *StructureInstance),
) {
	bm.onBuild = onBuild
	bm.onDestroy = onDestroy
}

// BuildStructure creates a new structure
func (bm *BaseManager) BuildStructure(structType StructureType, x, y float64, layer int, builder string) *StructureInstance {
	def := StructureRegistry[structType]
	if def == nil {
		return nil
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	structure := &StructureInstance{
		ID:       generateStructureID(),
		Def:      def,
		X:        x,
		Y:        y,
		Layer:    layer,
		Health:   def.Health,
		BuiltAt:  time.Now(),
		BuiltBy:  builder,
		IsActive: true,
	}

	bm.structures = append(bm.structures, structure)

	if bm.onBuild != nil {
		bm.onBuild(structure)
	}

	return structure
}

// GetStructureAt returns structure at position
func (bm *BaseManager) GetStructureAt(x, y float64, layer int, tolerance float64) *StructureInstance {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	tolSq := tolerance * tolerance

	for _, s := range bm.structures {
		if s.Layer != layer {
			continue
		}
		dx := s.X - x
		dy := s.Y - y
		if dx*dx+dy*dy <= tolSq {
			return s
		}
	}
	return nil
}

// GetStructuresInLayer returns all structures in layer
func (bm *BaseManager) GetStructuresInLayer(layer int) []*StructureInstance {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var result []*StructureInstance
	for _, s := range bm.structures {
		if s.Layer == layer {
			result = append(result, s)
		}
	}
	return result
}

// RemoveStructure removes a structure
func (bm *BaseManager) RemoveStructure(id string) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for i, s := range bm.structures {
		if s.ID == id {
			bm.structures = append(bm.structures[:i], bm.structures[i+1:]...)

			if bm.onDestroy != nil {
				bm.onDestroy(s)
			}
			return true
		}
	}
	return false
}

// DamageStructure damages a structure
func (bm *BaseManager) DamageStructure(id string, damage int) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for _, s := range bm.structures {
		if s.ID == id {
			s.TakeDamage(damage)
			if s.IsDestroyed() {
				bm.RemoveStructure(id)
			}
			return true
		}
	}
	return false
}

// RepairStructure repairs a structure
func (bm *BaseManager) RepairStructure(id string, amount int) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for _, s := range bm.structures {
		if s.ID == id {
			s.Repair(amount)
			return true
		}
	}
	return false
}

// GetBaseStats returns base statistics
func (bm *BaseManager) GetBaseStats(layer int) (total, walls, interactive int) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	for _, s := range bm.structures {
		if layer >= 0 && s.Layer != layer {
			continue
		}
		total++
		if s.Def.Type == STRUCT_WALL {
			walls++
		}
		if s.Def.IsInteractive {
			interactive++
		}
	}
	return
}

// ClearLayer removes all structures from a layer
func (bm *BaseManager) ClearLayer(layer int) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	var remaining []*StructureInstance
	for _, s := range bm.structures {
		if s.Layer != layer {
			remaining = append(remaining, s)
		} else if bm.onDestroy != nil {
			bm.onDestroy(s)
		}
	}
	bm.structures = remaining
}

// generateStructureID generates unique ID
func generateStructureID() string {
	return fmt.Sprintf("struct_%d", time.Now().UnixNano())
}
