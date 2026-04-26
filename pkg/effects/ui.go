package effects

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// EffectUI handles rendering of status effects
type EffectUI struct {
	screenWidth  int
	screenHeight int
	effectSize   float64
	spacing      float64
}

// NewEffectUI creates a new effect UI
func NewEffectUI(screenWidth, screenHeight int) *EffectUI {
	return &EffectUI{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		effectSize:   32,
		spacing:      4,
	}
}

// UpdateLayout updates the layout for new screen dimensions
func (ui *EffectUI) UpdateLayout(screenWidth, screenHeight int) {
	ui.screenWidth = screenWidth
	ui.screenHeight = screenHeight
}

// Draw renders active effects
func (ui *EffectUI) Draw(screen *ebiten.Image, effectManager *EffectManager) {
	effects := effectManager.GetActiveEffects()
	if len(effects) == 0 {
		return
	}

	// Position at top-left, below any other UI
	startX := 10.0
	startY := 150.0

	for i, effect := range effects {
		x := startX + float64(i)*(ui.effectSize+ui.spacing)
		y := startY

		ui.drawEffectIcon(screen, x, y, effect)
	}
}

// drawEffectIcon draws a single effect icon
func (ui *EffectUI) drawEffectIcon(screen *ebiten.Image, x, y float64, effect *ActiveEffect) {
	size := ui.effectSize

	// Determine color based on effect type
	bgColor := ui.getEffectColor(effect.Definition.Type)

	// Draw background (darker if negative)
	if effect.Definition.IsNegative {
		bgColor = color.RGBA{
			R: bgColor.(color.RGBA).R / 2,
			G: bgColor.(color.RGBA).G / 2,
			B: bgColor.(color.RGBA).B / 2,
			A: 200,
		}
	}

	// Draw icon background
	ebitenutil.DrawRect(screen, x, y, size, size, bgColor)

	// Draw border
	borderColor := color.RGBA{255, 255, 255, 150}
	if effect.Definition.IsNegative {
		borderColor = color.RGBA{255, 50, 50, 200}
	}
	thickness := 2.0
	ebitenutil.DrawRect(screen, x, y, size, thickness, borderColor)                // Top
	ebitenutil.DrawRect(screen, x, y+size-thickness, size, thickness, borderColor) // Bottom
	ebitenutil.DrawRect(screen, x, y, thickness, size, borderColor)                // Left
	ebitenutil.DrawRect(screen, x+size-thickness, y, thickness, size, borderColor) // Right

	// Draw stacks indicator if stacked
	if effect.Stacks > 1 {
		stackText := fmt.Sprintf("x%d", effect.Stacks)
		textX := int(x + size - 12)
		textY := int(y + size - 10)
		ebitenutil.DebugPrintAt(screen, stackText, textX, textY)
	}

	// Draw duration progress (overlay)
	progress := effect.GetProgress()
	remainingHeight := size * (1 - progress)
	overlayColor := color.RGBA{0, 0, 0, 100}
	ebitenutil.DrawRect(screen, x, y, size, size-remainingHeight, overlayColor)

	// Draw effect initial letter as icon
	if len(effect.Definition.Name) > 0 {
		letter := string(effect.Definition.Name[0])
		textX := int(x + size/2 - 4)
		textY := int(y + size/2 - 4)
		ebitenutil.DebugPrintAt(screen, letter, textX, textY)
	}
}

// getEffectColor returns a color for an effect type
func (ui *EffectUI) getEffectColor(effectType EffectType) color.Color {
	switch effectType {
	case EFFECT_POISON:
		return color.RGBA{50, 200, 50, 255}
	case EFFECT_REGENERATION:
		return color.RGBA{255, 100, 150, 255}
	case EFFECT_SPEED:
		return color.RGBA{255, 255, 100, 255}
	case EFFECT_SLOWNESS:
		return color.RGBA{150, 150, 100, 255}
	case EFFECT_STRENGTH:
		return color.RGBA{200, 50, 50, 255}
	case EFFECT_WEAKNESS:
		return color.RGBA{100, 100, 150, 255}
	case EFFECT_HUNGER:
		return color.RGBA{139, 90, 43, 255}
	case EFFECT_THIRST:
		return color.RGBA{50, 100, 220, 255}
	case EFFECT_BLEEDING:
		return color.RGBA{200, 30, 30, 255}
	case EFFECT_NIGHT_VISION:
		return color.RGBA{100, 100, 255, 255}
	case EFFECT_MINING_BOOST:
		return color.RGBA{255, 150, 50, 255}
	case EFFECT_DEFENSE:
		return color.RGBA{100, 100, 200, 255}
	default:
		return color.RGBA{150, 150, 150, 255}
	}
}

// DrawTooltip draws a tooltip for an effect
func (ui *EffectUI) DrawTooltip(screen *ebiten.Image, effect *ActiveEffect, mouseX, mouseY int) {
	if effect == nil {
		return
	}

	// Tooltip dimensions
	width := 200.0
	height := 80.0

	// Position tooltip near mouse but keep on screen
	x := float64(mouseX) + 10
	y := float64(mouseY) + 10

	if x+width > float64(ui.screenWidth) {
		x = float64(ui.screenWidth) - width - 10
	}
	if y+height > float64(ui.screenHeight) {
		y = float64(ui.screenHeight) - height - 10
	}

	// Draw tooltip background
	bgColor := color.RGBA{20, 20, 20, 230}
	if effect.Definition.IsNegative {
		bgColor = color.RGBA{50, 20, 20, 230}
	}
	ebitenutil.DrawRect(screen, x, y, width, height, bgColor)

	// Draw border
	borderColor := color.RGBA{150, 150, 150, 255}
	if effect.Definition.IsNegative {
		borderColor = color.RGBA{200, 50, 50, 255}
	}
	thickness := 2.0
	ebitenutil.DrawRect(screen, x, y, width, thickness, borderColor)
	ebitenutil.DrawRect(screen, x, y+height-thickness, width, thickness, borderColor)
	ebitenutil.DrawRect(screen, x, y, thickness, height, borderColor)
	ebitenutil.DrawRect(screen, x+width-thickness, y, thickness, height, borderColor)

	// Draw effect name
	textY := int(y + 8)
	ebitenutil.DebugPrintAt(screen, effect.Definition.Name, int(x+8), textY)

	// Draw description
	textY += 20
	ebitenutil.DebugPrintAt(screen, effect.Definition.Description, int(x+8), textY)

	// Draw remaining time
	remaining := effect.GetRemainingDuration()
	seconds := int(remaining.Seconds())
	minutes := seconds / 60
	seconds = seconds % 60
	timeText := fmt.Sprintf("Time: %d:%02d", minutes, seconds)
	textY += 20
	ebitenutil.DebugPrintAt(screen, timeText, int(x+8), textY)

	// Draw stacks if applicable
	if effect.Stacks > 1 {
		stackText := fmt.Sprintf("Stacks: %d/%d", effect.Stacks, effect.Definition.MaxStacks)
		textY += 16
		ebitenutil.DebugPrintAt(screen, stackText, int(x+8), textY)
	}
}

// GetEffectAtPosition returns the effect at a screen position
func (ui *EffectUI) GetEffectAtPosition(effectManager *EffectManager, mouseX, mouseY int) *ActiveEffect {
	effects := effectManager.GetActiveEffects()
	if len(effects) == 0 {
		return nil
	}

	startX := 10.0
	startY := 150.0

	for _, effect := range effects {
		x := startX
		y := startY

		if float64(mouseX) >= x && float64(mouseX) <= x+ui.effectSize &&
			float64(mouseY) >= y && float64(mouseY) <= y+ui.effectSize {
			return effect
		}

		startX += ui.effectSize + ui.spacing
	}

	return nil
}

// DrawEffectBar draws a progress bar for an effect
func (ui *EffectUI) DrawEffectBar(screen *ebiten.Image, x, y, width, height float64, effect *ActiveEffect, barColor color.Color) {
	// Background
	bgColor := color.RGBA{30, 30, 30, 200}
	ebitenutil.DrawRect(screen, x, y, width, height, bgColor)

	// Fill based on remaining duration
	progress := 1.0 - effect.GetProgress()
	fillWidth := width * progress

	// Pulse effect for critical (low duration)
	finalColor := barColor
	if progress < 0.2 {
		pulse := uint8(50 * math.Sin(float64(ebiten.CurrentTPS())*0.1))
		rgba, ok := barColor.(color.RGBA)
		if ok {
			finalColor = color.RGBA{
				R: rgba.R + pulse,
				G: rgba.G,
				B: rgba.B,
				A: rgba.A,
			}
		}
	}

	ebitenutil.DrawRect(screen, x, y, fillWidth, height, finalColor)
}
