package gathering

import (
	"fmt"
	"math/rand"
	"time"
)

// GatherableType represents gatherable resources
type GatherableType int

const (
	GATHER_BERRIES GatherableType = iota
	GATHER_MUSHROOM
	GATHER_HERBS
	GATHER_WOOD
	GATHER_STONE_SURFACE
)

// ResourceNode represents a gatherable node
type ResourceNode struct {
	ID           string
	Type         GatherableType
	X, Y         float64
	Layer        int
	Yield        map[string]int // Item ID -> quantity
	RegenTime    time.Duration
	LastGathered *time.Time
	Respawns     bool
}

// CanGather checks if node can be gathered
func (rn *ResourceNode) CanGather() bool {
	if !rn.Respawns {
		return rn.LastGathered == nil
	}

	if rn.LastGathered == nil {
		return true
	}

	return time.Since(*rn.LastGathered) >= rn.RegenTime
}

// Gather gathers from the node
func (rn *ResourceNode) Gather() map[string]int {
	if !rn.CanGather() {
		return nil
	}

	now := time.Now()
	rn.LastGathered = &now

	// Add some randomness to yields
	result := make(map[string]int)
	for item, qty := range rn.Yield {
		actualQty := qty + rand.Intn(qty/2+1) - qty/4 // +/- 25%
		if actualQty < 1 {
			actualQty = 1
		}
		result[item] = actualQty
	}

	return result
}

// GenerateNodes creates resource nodes in an area
func GenerateNodes(count int, worldWidth, worldHeight float64, layer int) []*ResourceNode {
	var nodes []*ResourceNode

	types := []GatherableType{GATHER_BERRIES, GATHER_MUSHROOM, GATHER_HERBS}

	for i := 0; i < count; i++ {
		nodeType := types[rand.Intn(len(types))]

		node := &ResourceNode{
			ID:        generateNodeID(i),
			Type:      nodeType,
			X:         rand.Float64() * worldWidth,
			Y:         rand.Float64() * worldHeight,
			Layer:     layer,
			Respawns:  true,
			RegenTime: 5 * time.Minute,
			Yield:     getYieldForType(nodeType),
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func getYieldForType(t GatherableType) map[string]int {
	switch t {
	case GATHER_BERRIES:
		return map[string]int{"berries": 2, "seeds": 1}
	case GATHER_MUSHROOM:
		return map[string]int{"mushroom": 1, "spores": 1}
	case GATHER_HERBS:
		return map[string]int{"herbs": 2}
	default:
		return map[string]int{"resource": 1}
	}
}

func generateNodeID(index int) string {
	return fmt.Sprintf("node_%d_%d", time.Now().Unix(), index)
}
