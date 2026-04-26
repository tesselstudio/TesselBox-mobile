package achievements

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// AchievementUI handles achievement display
type AchievementUI struct {
	screenWidth   int
	screenHeight  int
	visible       bool
	selectedTab   AchievementCategory
	scrollOffset  int
	popupQueue    []*AchievementProgress
	popupTimer    float64
	popupDuration float64
}

// NewAchievementUI creates a new achievement UI
func NewAchievementUI(screenWidth, screenHeight int) *AchievementUI {
	return &AchievementUI{
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		visible:       false,
		selectedTab:   CATEGORY_GENERAL,
		popupQueue:    make([]*AchievementProgress, 0),
		popupDuration: 5.0, // Show popup for 5 seconds
	}
}

// Toggle toggles the achievement UI visibility
func (ui *AchievementUI) Toggle() {
	ui.visible = !ui.visible
}

// IsVisible returns true if UI is visible
func (ui *AchievementUI) IsVisible() bool {
	return ui.visible
}

// Show shows the achievement UI
func (ui *AchievementUI) Show() {
	ui.visible = true
}

// Hide hides the achievement UI
func (ui *AchievementUI) Hide() {
	ui.visible = false
}

// QueuePopup adds an achievement to the popup queue
func (ui *AchievementUI) QueuePopup(progress *AchievementProgress) {
	ui.popupQueue = append(ui.popupQueue, progress)
}

// Update handles input and animations
func (ui *AchievementUI) Update(deltaTime float64) error {
	// Handle popup timer
	if len(ui.popupQueue) > 0 {
		ui.popupTimer += deltaTime
		if ui.popupTimer >= ui.popupDuration {
			// Remove current popup
			if len(ui.popupQueue) > 1 {
				ui.popupQueue = ui.popupQueue[1:]
			} else {
				ui.popupQueue = make([]*AchievementProgress, 0)
			}
			ui.popupTimer = 0
		}
	}

	if !ui.visible {
		return nil
	}

	// Handle tab switching
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		ui.selectedTab = (ui.selectedTab + 1) % 7
		ui.scrollOffset = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		ui.selectedTab = (ui.selectedTab + 6) % 7
		ui.scrollOffset = 0
	}

	// Handle scrolling
	_, scrollY := ebiten.Wheel()
	if scrollY > 0 {
		ui.scrollOffset = maxInt(0, ui.scrollOffset-1)
	} else if scrollY < 0 {
		ui.scrollOffset++
	}

	// Handle escape
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.Hide()
	}

	return nil
}

// Draw renders the achievement UI
func (ui *AchievementUI) Draw(screen *ebiten.Image, manager *AchievementManager) {
	// Draw popup if active
	if len(ui.popupQueue) > 0 {
		ui.drawPopup(screen, ui.popupQueue[0])
	}

	if !ui.visible {
		return
	}

	// Draw background
	bgColor := color.RGBA{20, 20, 30, 240}
	ebitenutil.DrawRect(screen, 0, 0, float64(ui.screenWidth), float64(ui.screenHeight), bgColor)

	// Draw title
	titleX := ui.screenWidth/2 - 60
	ebitenutil.DebugPrintAt(screen, "ACHIEVEMENTS", titleX, 20)

	// Draw completion percentage
	completion := manager.GetCompletionPercentage()
	completionText := fmt.Sprintf("Completion: %.1f%%", completion)
	ebitenutil.DebugPrintAt(screen, completionText, ui.screenWidth-150, 20)

	// Draw tabs
	ui.drawTabs(screen)

	// Draw achievement list
	ui.drawAchievementList(screen, manager)
}

// drawTabs renders category tabs
func (ui *AchievementUI) drawTabs(screen *ebiten.Image) {
	tabWidth := 120
	tabHeight := 30
	startX := 20
	y := 60

	categories := []AchievementCategory{
		CATEGORY_GENERAL,
		CATEGORY_MINING,
		CATEGORY_CRAFTING,
		CATEGORY_SURVIVAL,
		CATEGORY_EXPLORATION,
		CATEGORY_COMBAT,
		CATEGORY_BUILDING,
	}

	for i, cat := range categories {
		x := startX + i*(tabWidth+10)

		// Draw tab background
		bgColor := color.RGBA{50, 50, 60, 255}
		if cat == ui.selectedTab {
			bgColor = color.RGBA{80, 80, 120, 255}
		}
		ebitenutil.DrawRect(screen, float64(x), float64(y), float64(tabWidth), float64(tabHeight), bgColor)

		// Draw tab text
		text := cat.String()
		textX := x + (tabWidth-len(text)*8)/2
		ebitenutil.DebugPrintAt(screen, text, textX, y+8)
	}
}

// drawAchievementList renders achievements for selected category
func (ui *AchievementUI) drawAchievementList(screen *ebiten.Image, manager *AchievementManager) {
	achievements := manager.GetProgressByCategory(ui.selectedTab)

	itemHeight := 70
	itemWidth := ui.screenWidth - 60
	startX := 30
	startY := 110
	visibleCount := (ui.screenHeight - startY - 20) / itemHeight

	for i, progress := range achievements {
		if i < ui.scrollOffset {
			continue
		}
		if i >= ui.scrollOffset+visibleCount {
			break
		}

		localIndex := i - ui.scrollOffset
		y := startY + localIndex*itemHeight

		ui.drawAchievementItem(screen, progress, startX, y, itemWidth, itemHeight-10)
	}

	// Draw scroll indicator if needed
	if len(achievements) > visibleCount {
		ui.drawScrollIndicator(screen, ui.scrollOffset, len(achievements), visibleCount)
	}
}

// drawAchievementItem renders a single achievement entry
func (ui *AchievementUI) drawAchievementItem(screen *ebiten.Image, progress *AchievementProgress, x, y, width, height int) {
	def := progress.Definition

	// Background color based on tier and unlock status
	bgColor := ui.getTierColor(def.Tier)
	if !progress.Unlocked {
		// Darken locked achievements
		bgColor = color.RGBA{
			R: bgColor.(color.RGBA).R / 3,
			G: bgColor.(color.RGBA).G / 3,
			B: bgColor.(color.RGBA).B / 3,
			A: 150,
		}
	}

	// Draw background
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(width), float64(height), bgColor)

	// Draw border
	borderColor := color.RGBA{150, 150, 150, 255}
	if progress.Unlocked {
		borderColor = color.RGBA{255, 255, 255, 255}
	}
	thickness := 2.0
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(width), thickness, borderColor)
	ebitenutil.DrawRect(screen, float64(x), float64(y+height)-thickness, float64(width), thickness, borderColor)
	ebitenutil.DrawRect(screen, float64(x), float64(y), thickness, float64(height), borderColor)
	ebitenutil.DrawRect(screen, float64(x+width)-thickness, float64(y), thickness, float64(height), borderColor)

	// Draw icon placeholder (colored square)
	iconSize := 48
	iconX := x + 10
	iconY := y + (height-iconSize)/2
	iconColor := ui.getTierColor(def.Tier)
	ebitenutil.DrawRect(screen, float64(iconX), float64(iconY), float64(iconSize), float64(iconSize), iconColor)

	// Draw name
	nameX := iconX + iconSize + 15
	nameY := y + 10
	ebitenutil.DebugPrintAt(screen, def.Name, nameX, nameY)

	// Draw description
	descY := nameY + 20
	descColor := color.RGBA{200, 200, 200, 255}
	if !progress.Unlocked {
		descColor = color.RGBA{100, 100, 100, 255}
	}
	// We need to use the color, so let's draw with it
	_ = descColor
	ebitenutil.DebugPrintAt(screen, def.Description, nameX, descY)

	// Draw progress bar for progress-based achievements
	if def.IsProgressBased() && !progress.Unlocked {
		barY := descY + 20
		barWidth := 200
		barHeight := 10
		progressPct := progress.GetProgressPercentage() / 100.0

		// Background
		ebitenutil.DrawRect(screen, float64(nameX), float64(barY), float64(barWidth), float64(barHeight), color.RGBA{50, 50, 50, 255})
		// Fill
		fillWidth := float64(barWidth) * progressPct
		ebitenutil.DrawRect(screen, float64(nameX), float64(barY), fillWidth, float64(barHeight), color.RGBA{100, 200, 100, 255})
		// Text
		progressText := fmt.Sprintf("%d/%d", progress.Progress, def.MaxProgress)
		ebitenutil.DebugPrintAt(screen, progressText, nameX+barWidth+10, barY)
	}

	// Draw reward info
	rewardText := fmt.Sprintf("+%d XP", def.RewardXP)
	rewardX := x + width - 80
	rewardY := y + 10
	ebitenutil.DebugPrintAt(screen, rewardText, rewardX, rewardY)

	// Draw tier indicator
	tierText := def.Tier.String()
	tierX := x + width - 80
	tierY := y + height - 20
	ebitenutil.DebugPrintAt(screen, tierText, tierX, tierY)

	// Draw unlocked timestamp
	if progress.Unlocked && progress.UnlockedAt != nil {
		unlockedText := "UNLOCKED"
		unlockedX := x + width - 100
		unlockedY := y + height/2 - 5
		ebitenutil.DebugPrintAt(screen, unlockedText, unlockedX, unlockedY)
	}
}

// drawPopup renders an achievement unlock notification
func (ui *AchievementUI) drawPopup(screen *ebiten.Image, progress *AchievementProgress) {
	def := progress.Definition

	// Popup dimensions and position (bottom-right corner)
	popupWidth := 300.0
	popupHeight := 80.0
	x := float64(ui.screenWidth) - popupWidth - 20
	y := float64(ui.screenHeight) - popupHeight - 20

	// Fade in/out animation
	alpha := uint8(255)
	if ui.popupTimer < 0.5 {
		// Fade in
		alpha = uint8(ui.popupTimer / 0.5 * 255)
	} else if ui.popupTimer > ui.popupDuration-0.5 {
		// Fade out
		alpha = uint8((ui.popupDuration - ui.popupTimer) / 0.5 * 255)
	}

	// Draw background
	bgColor := ui.getTierColor(def.Tier)
	if rgba, ok := bgColor.(color.RGBA); ok {
		bgColor = color.RGBA{rgba.R, rgba.G, rgba.B, alpha}
	}
	ebitenutil.DrawRect(screen, x, y, popupWidth, popupHeight, bgColor)

	// Draw border
	borderColor := color.RGBA{255, 255, 255, alpha}
	thickness := 3.0
	ebitenutil.DrawRect(screen, x, y, popupWidth, thickness, borderColor)
	ebitenutil.DrawRect(screen, x, y+popupHeight-thickness, popupWidth, thickness, borderColor)
	ebitenutil.DrawRect(screen, x, y, thickness, popupHeight, borderColor)
	ebitenutil.DrawRect(screen, x+popupWidth-thickness, y, thickness, popupHeight, borderColor)

	// Draw icon placeholder
	iconSize := 48.0
	iconX := x + 15
	iconY := y + (popupHeight-iconSize)/2
	iconColor := color.RGBA{255, 255, 255, alpha}
	ebitenutil.DrawRect(screen, iconX, iconY, iconSize, iconSize, iconColor)

	// Draw text
	textX := iconX + iconSize + 15
	textY := y + 15

	// Title
	ebitenutil.DebugPrintAt(screen, "ACHIEVEMENT UNLOCKED!", int(textX), int(textY))

	// Achievement name
	textY += 25
	ebitenutil.DebugPrintAt(screen, def.Name, int(textX), int(textY))

	// Reward
	textY += 20
	rewardText := fmt.Sprintf("+%d XP", def.RewardXP)
	ebitenutil.DebugPrintAt(screen, rewardText, int(textX), int(textY))
}

// drawScrollIndicator renders scroll position indicator
func (ui *AchievementUI) drawScrollIndicator(screen *ebiten.Image, offset, total, visible int) {
	barX := float64(ui.screenWidth - 20)
	barY := 110.0
	barWidth := 8.0
	barHeight := float64(ui.screenHeight - 140)

	// Background
	ebitenutil.DrawRect(screen, barX, barY, barWidth, barHeight, color.RGBA{50, 50, 50, 200})

	// Scroll thumb
	thumbHeight := barHeight * float64(visible) / float64(total)
	thumbY := barY + (barHeight-thumbHeight)*float64(offset)/float64(total-visible)
	ebitenutil.DrawRect(screen, barX, thumbY, barWidth, thumbHeight, color.RGBA{150, 150, 150, 255})
}

// getTierColor returns color for achievement tier
func (ui *AchievementUI) getTierColor(tier AchievementTier) color.Color {
	switch tier {
	case TIER_BRONZE:
		return color.RGBA{205, 127, 50, 255}
	case TIER_SILVER:
		return color.RGBA{192, 192, 192, 255}
	case TIER_GOLD:
		return color.RGBA{255, 215, 0, 255}
	case TIER_PLATINUM:
		return color.RGBA{229, 228, 226, 255}
	case TIER_DIAMOND:
		return color.RGBA{185, 242, 255, 255}
	default:
		return color.RGBA{150, 150, 150, 255}
	}
}

// UpdateLayout updates layout for new screen size
func (ui *AchievementUI) UpdateLayout(screenWidth, screenHeight int) {
	ui.screenWidth = screenWidth
	ui.screenHeight = screenHeight
}

// Helper function
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
