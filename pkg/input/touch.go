package input

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// TouchZone defines screen regions for touch controls
type TouchZone int

const (
	ZoneLeftHalf  TouchZone = iota // Movement zone (left 50% of screen)
	ZoneRightHalf                  // Action zone (right 50% of screen)
)

// TouchGesture represents detected touch gestures
type TouchGesture int

const (
	GestureNone TouchGesture = iota
	GestureTap
	GestureSwipeUp
	GestureSwipeDown
	GestureSwipeLeft
	GestureSwipeRight
	GesturePinchIn
	GesturePinchOut
	GestureHold
)

// TouchInfo holds information about a single touch
type TouchInfo struct {
	ID        ebiten.TouchID
	StartX    float64
	StartY    float64
	CurrentX  float64
	CurrentY  float64
	StartTime time.Time
	IsActive  bool
	Zone      TouchZone
}

// TouchInputManager handles touch gestures for mobile controls
type TouchInputManager struct {
	// Active touch tracking
	touches map[ebiten.TouchID]*TouchInfo

	// Gesture detection results (cleared each frame)
	gestures map[TouchZone]TouchGesture

	// Swipe detection threshold
	swipeThreshold float64

	// Hold detection threshold
	holdThreshold time.Duration

	// Screen dimensions for zone calculation
	screenWidth  float64
	screenHeight float64

	// Pinch tracking
	lastPinchDistance float64
	pinchActive       bool

	// Gesture state for this frame
	moveDirectionX float64 // -1 to 1 (left to right)
	moveDirectionY float64 // -1 to 1 (up to down)
	isMining       bool
	isPlacing      bool
	isJumping      bool
	zoomDelta      float64 // positive = zoom in, negative = zoom out
}

// NewTouchInputManager creates a new touch input manager
func NewTouchInputManager() *TouchInputManager {
	return &TouchInputManager{
		touches:        make(map[ebiten.TouchID]*TouchInfo),
		gestures:       make(map[TouchZone]TouchGesture),
		swipeThreshold: 30.0,                   // pixels to count as a swipe
		holdThreshold:  500 * time.Millisecond, // 500ms for hold
		screenWidth:    800,
		screenHeight:   600,
		moveDirectionX: 0,
		moveDirectionY: 0,
		isMining:       false,
		isPlacing:      false,
		isJumping:      false,
		zoomDelta:      0,
	}
}

// Update processes all touch events for this frame
func (tim *TouchInputManager) Update() {
	// Reset gesture state
	tim.moveDirectionX = 0
	tim.moveDirectionY = 0
	tim.isMining = false
	tim.isPlacing = false
	tim.isJumping = false
	tim.zoomDelta = 0

	// Get current screen dimensions
	w, h := ebiten.WindowSize()
	tim.screenWidth, tim.screenHeight = float64(w), float64(h)

	// Process new touches
	justPressedIDs := inpututil.AppendJustPressedTouchIDs(nil)
	for _, id := range justPressedIDs {
		x, y := ebiten.TouchPosition(id)
		zone := tim.getZoneForPosition(x, y)
		tim.touches[id] = &TouchInfo{
			ID:        id,
			StartX:    float64(x),
			StartY:    float64(y),
			CurrentX:  float64(x),
			CurrentY:  float64(y),
			StartTime: time.Now(),
			IsActive:  true,
			Zone:      zone,
		}
	}

	// Update active touches
	activeIDs := ebiten.AppendTouchIDs(nil)
	activeMap := make(map[ebiten.TouchID]bool)
	for _, id := range activeIDs {
		activeMap[id] = true
		if touch, exists := tim.touches[id]; exists {
			x, y := ebiten.TouchPosition(id)
			touch.CurrentX = float64(x)
			touch.CurrentY = float64(y)
		}
	}

	// Handle pinch gesture (two fingers)
	if len(activeIDs) == 2 {
		tim.processPinchGesture(activeIDs)
	} else {
		tim.pinchActive = false
	}

	// Process released touches to detect gestures
	justReleasedIDs := inpututil.AppendJustReleasedTouchIDs(nil)
	for _, id := range justReleasedIDs {
		if touch, exists := tim.touches[id]; exists {
			tim.processTouchRelease(touch)
			delete(tim.touches, id)
		}
	}

	// Process ongoing touches for continuous actions
	for _, touch := range tim.touches {
		tim.processOngoingTouch(touch)
	}
}

// getZoneForPosition determines which zone a touch is in
func (tim *TouchInputManager) getZoneForPosition(x, y int) TouchZone {
	if float64(x) < tim.screenWidth/2 {
		return ZoneLeftHalf
	}
	return ZoneRightHalf
}

// processPinchGesture handles two-finger zoom
func (tim *TouchInputManager) processPinchGesture(ids []ebiten.TouchID) {
	if len(ids) != 2 {
		return
	}

	touch1 := tim.touches[ids[0]]
	touch2 := tim.touches[ids[1]]
	if touch1 == nil || touch2 == nil {
		return
	}

	// Calculate distance between touch points
	dx := touch1.CurrentX - touch2.CurrentX
	dy := touch1.CurrentY - touch2.CurrentY
	distance := math.Sqrt(dx*dx + dy*dy)

	if tim.pinchActive {
		// Calculate zoom delta
		delta := distance - tim.lastPinchDistance
		tim.zoomDelta = delta * 0.01 // Scale down the delta
	}

	tim.lastPinchDistance = distance
	tim.pinchActive = true
}

// processTouchRelease handles gestures when a touch ends
func (tim *TouchInputManager) processTouchRelease(touch *TouchInfo) {
	duration := time.Since(touch.StartTime)
	dx := touch.CurrentX - touch.StartX
	dy := touch.CurrentY - touch.StartY
	absDx := math.Abs(dx)
	absDy := math.Abs(dy)

	// Determine if it's a swipe or tap
	isSwipe := absDx > tim.swipeThreshold || absDy > tim.swipeThreshold

	if isSwipe {
		// Determine swipe direction
		if absDx > absDy {
			// Horizontal swipe
			if dx > 0 {
				tim.gestures[touch.Zone] = GestureSwipeRight
			} else {
				tim.gestures[touch.Zone] = GestureSwipeLeft
			}
		} else {
			// Vertical swipe
			if dy > 0 {
				tim.gestures[touch.Zone] = GestureSwipeDown
			} else {
				tim.gestures[touch.Zone] = GestureSwipeUp
			}
		}
	} else if duration < tim.holdThreshold {
		// It's a tap
		tim.gestures[touch.Zone] = GestureTap
	}

	// Apply gestures to actions
	switch touch.Zone {
	case ZoneLeftHalf:
		// Left side - movement controls
		if isSwipe {
			if absDx > absDy {
				if dx > 0 {
					tim.moveDirectionX = 1 // Right
				} else {
					tim.moveDirectionX = -1 // Left
				}
			} else {
				if dy < 0 {
					tim.isJumping = true
				}
			}
		}
	case ZoneRightHalf:
		// Right side - action controls
		if !isSwipe && duration < tim.holdThreshold {
			// Quick tap = mine
			tim.isMining = true
		}
	}
}

// processOngoingTouch handles continuous touch actions (like holding for placement)
func (tim *TouchInputManager) processOngoingTouch(touch *TouchInfo) {
	duration := time.Since(touch.StartTime)
	dx := touch.CurrentX - touch.StartX
	dy := touch.CurrentY - touch.StartY
	absDx := math.Abs(dx)
	absDy := math.Abs(dy)

	switch touch.Zone {
	case ZoneLeftHalf:
		// Left side - continuous movement while dragging
		if absDx > tim.swipeThreshold {
			if dx > 0 {
				tim.moveDirectionX = 1
			} else {
				tim.moveDirectionX = -1
			}
		}
		if absDy > tim.swipeThreshold {
			if dy > 0 {
				tim.moveDirectionY = 1 // Down
			} else {
				tim.moveDirectionY = -1 // Up
			}
		}

	case ZoneRightHalf:
		// Right side - hold for placement, continuous mine while dragging
		if duration > tim.holdThreshold && absDx < tim.swipeThreshold && absDy < tim.swipeThreshold {
			// Long hold without movement = place block
			tim.isPlacing = true
		} else {
			// Continuous mining while touching right side
			tim.isMining = true
		}
	}
}

// Movement accessors
func (tim *TouchInputManager) GetMoveDirectionX() float64 {
	return tim.moveDirectionX
}

func (tim *TouchInputManager) GetMoveDirectionY() float64 {
	return tim.moveDirectionY
}

func (tim *TouchInputManager) IsJumping() bool {
	return tim.isJumping
}

// Action accessors
func (tim *TouchInputManager) IsMining() bool {
	return tim.isMining
}

func (tim *TouchInputManager) IsPlacing() bool {
	return tim.isPlacing
}

// Zoom accessor
func (tim *TouchInputManager) GetZoomDelta() float64 {
	return tim.zoomDelta
}

// IsTouchActive returns true if any touch is currently active
func (tim *TouchInputManager) IsTouchActive() bool {
	return len(ebiten.AppendTouchIDs(nil)) > 0
}

// TouchPosition returns the position of a specific touch
func (tim *TouchInputManager) TouchPosition(id ebiten.TouchID) (float64, float64) {
	x, y := ebiten.TouchPosition(id)
	return float64(x), float64(y)
}

// ActiveTouchCount returns the number of active touches
func (tim *TouchInputManager) ActiveTouchCount() int {
	return len(ebiten.AppendTouchIDs(nil))
}
