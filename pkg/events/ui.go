package events

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// EventUI handles event notifications and display
type EventUI struct {
	screenWidth   int
	screenHeight  int
	showWarning   bool
	warningTimer  float64
	warningFadeIn float64
	currentEvent  *ActiveEvent
}

// NewEventUI creates a new event UI
func NewEventUI(screenWidth, screenHeight int) *EventUI {
	return &EventUI{
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		warningFadeIn: 1.0,
	}
}

// UpdateLayout updates layout for new screen size
func (ui *EventUI) UpdateLayout(screenWidth, screenHeight int) {
	ui.screenWidth = screenWidth
	ui.screenHeight = screenHeight
}

// OnEventWarning called when an event warning occurs
func (ui *EventUI) OnEventWarning(event *ActiveEvent) {
	ui.currentEvent = event
	ui.showWarning = true
	ui.warningTimer = 0
	ui.warningFadeIn = 0
}

// OnEventStart called when an event starts
func (ui *EventUI) OnEventStart(event *ActiveEvent) {
	ui.currentEvent = event
	ui.showWarning = false
}

// OnEventEnd called when an event ends
func (ui *EventUI) OnEventEnd(event *ActiveEvent) {
	if ui.currentEvent == event {
		ui.currentEvent = nil
		ui.showWarning = false
	}
}

// Update updates UI animations
func (ui *EventUI) Update(deltaTime float64) {
	// Fade in warning
	if ui.showWarning && ui.warningFadeIn < 1.0 {
		ui.warningFadeIn += deltaTime * 2
		if ui.warningFadeIn > 1.0 {
			ui.warningFadeIn = 1.0
		}
	}

	// Pulse effect for warning
	if ui.showWarning {
		ui.warningTimer += deltaTime * 3
	}
}

// Draw renders event UI elements
func (ui *EventUI) Draw(screen *ebiten.Image) {
	// Draw warning banner
	if ui.showWarning && ui.currentEvent != nil {
		ui.DrawWarningBanner(screen)
	}

	// Draw active event indicator
	if ui.currentEvent != nil && ui.currentEvent.Phase == PHASE_ACTIVE {
		ui.DrawActiveEventIndicator(screen)
	}
}

// drawWarningBanner renders the event warning
func (ui *EventUI) DrawWarningBanner(screen *ebiten.Image) {
	bannerHeight := 60.0
	y := 80.0

	// Pulsing alpha
	pulse := uint8(200 + uint8(55*math.Sin(ui.warningTimer)))
	alpha := uint8(float64(pulse) * ui.warningFadeIn)

	// Background
	bgColor := color.RGBA{200, 150, 50, alpha}
	if ui.currentEvent.Definition.Severity >= SEVERITY_SEVERE {
		bgColor = color.RGBA{200, 50, 50, alpha}
	}
	ebitenutil.DrawRect(screen, 0, y, float64(ui.screenWidth), bannerHeight, bgColor)

	// Border
	borderColor := color.RGBA{255, 255, 255, alpha}
	ebitenutil.DrawRect(screen, 0, y, float64(ui.screenWidth), 3, borderColor)
	ebitenutil.DrawRect(screen, 0, y+bannerHeight-3, float64(ui.screenWidth), 3, borderColor)

	// Text
	textY := int(y + 20)
	def := ui.currentEvent.Definition

	// Warning label
	warningText := "WARNING"
	textX := ui.screenWidth/2 - len(warningText)*4
	ebitenutil.DebugPrintAt(screen, warningText, textX, textY)

	// Event name and time
	textY += 20
	timeUntil := ui.currentEvent.GetTimeUntilStart()
	seconds := int(timeUntil.Seconds())
	message := fmt.Sprintf("%s in %ds! %s", def.Name, seconds, def.WarningMessage)
	textX = ui.screenWidth/2 - len(message)*4
	ebitenutil.DebugPrintAt(screen, message, textX, textY)
}

// drawActiveEventIndicator renders the active event status
func (ui *EventUI) DrawActiveEventIndicator(screen *ebiten.Image) {
	if ui.currentEvent == nil {
		return
	}

	def := ui.currentEvent.Definition

	// Box dimensions
	boxWidth := 250.0
	boxHeight := 60.0
	x := float64(ui.screenWidth) - boxWidth - 20
	y := 150.0

	// Background based on severity
	bgColor := ui.getSeverityColor(def.Severity)
	ebitenutil.DrawRect(screen, x, y, boxWidth, boxHeight, bgColor)

	// Border
	borderColor := color.RGBA{255, 255, 255, 255}
	thickness := 2.0
	ebitenutil.DrawRect(screen, x, y, boxWidth, thickness, borderColor)
	ebitenutil.DrawRect(screen, x, y+boxHeight-thickness, boxWidth, thickness, borderColor)
	ebitenutil.DrawRect(screen, x, y, thickness, boxHeight, borderColor)
	ebitenutil.DrawRect(screen, x+boxWidth-thickness, y, thickness, boxHeight, borderColor)

	// Event name
	textX := int(x + 10)
	textY := int(y + 10)
	ebitenutil.DebugPrintAt(screen, def.Name, textX, textY)

	// Progress bar
	barY := y + 35
	barWidth := boxWidth - 20
	barHeight := 15.0
	progress := ui.currentEvent.GetProgress()

	// Bar background
	ebitenutil.DrawRect(screen, x+10, barY, barWidth, barHeight, color.RGBA{50, 50, 50, 255})

	// Bar fill
	fillWidth := barWidth * (1 - progress)
	fillColor := color.RGBA{200, 200, 50, 255}
	if def.Severity >= SEVERITY_SEVERE {
		fillColor = color.RGBA{200, 50, 50, 255}
	}
	ebitenutil.DrawRect(screen, x+10, barY, fillWidth, barHeight, fillColor)

	// Time remaining
	remaining := ui.currentEvent.GetRemainingTime()
	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60
	timeText := fmt.Sprintf("%d:%02d", minutes, seconds)
	textY = int(barY + 2)
	textX = int(x+10+barWidth/2) - len(timeText)*4
	ebitenutil.DebugPrintAt(screen, timeText, textX, textY)
}

// getSeverityColor returns color based on severity
func (ui *EventUI) getSeverityColor(severity EventSeverity) color.Color {
	switch severity {
	case SEVERITY_MILD:
		return color.RGBA{100, 200, 100, 200}
	case SEVERITY_MODERATE:
		return color.RGBA{200, 200, 100, 200}
	case SEVERITY_SEVERE:
		return color.RGBA{200, 100, 50, 200}
	case SEVERITY_EXTREME:
		return color.RGBA{200, 50, 50, 200}
	default:
		return color.RGBA{150, 150, 150, 200}
	}
}

// IsShowingWarning returns true if warning is currently displayed
func (ui *EventUI) IsShowingWarning() bool {
	return ui.showWarning
}

// GetCurrentEvent returns the current active event
func (ui *EventUI) GetCurrentEvent() *ActiveEvent {
	return ui.currentEvent
}
