package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/tesselstudio/TesselBox-mobile/pkg/blocks"
	"github.com/tesselstudio/TesselBox-mobile/pkg/game"
	"github.com/tesselstudio/TesselBox-mobile/pkg/items"
	"github.com/tesselstudio/TesselBox-mobile/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// DrawingExample shows how to adapt the drawing logic to use GameManager
// This is a reference implementation for migrating the Draw() method

// DrawGameScene draws the main game world
func DrawGameScene(screen *ebiten.Image, gm *game.GameManager) {
	// Clear screen
	screen.Fill(color.RGBA{135, 206, 235, 255}) // Sky blue

	// Draw world using camera position
	drawWorld(screen, gm)

	// Draw UI overlay
	drawGameUI(screen, gm)
}

// drawWorld draws the game world with camera offset
func drawWorld(screen *ebiten.Image, gm *game.GameManager) {
	// Use camera position from GameManager
	cameraX := gm.CameraX
	cameraY := gm.CameraY

	// Draw hexagons/blocks
	if gm.World != nil {
		// Get visible chunks based on camera position
		gm.World.GetChunksInRange(cameraX, cameraY)

		// Iterate over all loaded chunks
		for _, chunk := range gm.World.Chunks {
			if chunk == nil {
				continue
			}

			// Draw each hexagon in the chunk
			for _, hex := range chunk.Hexagons {
				if hex == nil {
					continue
				}

				// Calculate screen position
				screenX := hex.X - cameraX
				screenY := hex.Y - cameraY

				// Only draw if on screen
				if screenX > -100 && screenX < float64(ScreenWidth)+100 &&
					screenY > -100 && screenY < float64(ScreenHeight)+100 {
					// Draw hexagon (implementation depends on your rendering system)
					drawHexagon(screen, screenX, screenY, hex, gm)
				}
			}
		}
	}

	// Draw player
	if gm.Player != nil {
		playerScreenX := gm.Player.X - cameraX
		playerScreenY := gm.Player.Y - cameraY
		drawPlayer(screen, playerScreenX, playerScreenY, gm.Player)
	}

	// Draw dropped items
	for _, item := range gm.DroppedItems {
		itemScreenX := item.X - cameraX
		itemScreenY := item.Y - cameraY
		drawDroppedItem(screen, itemScreenX, itemScreenY, item)
	}
}

// drawGameUI draws the game UI elements
func drawGameUI(screen *ebiten.Image, gm *game.GameManager) {
	// Draw hotbar
	if gm.Inventory != nil {
		drawHotbar(screen, gm.Inventory)
	}

	// Draw health bar
	if gm.HealthSystem != nil {
		drawHealthBar(screen, gm.HealthSystem)
	}

	// Draw HUD
	if gm.HUD != nil {
		gm.HUD.Draw(screen)
	}

	// Draw debug info if enabled
	drawDebugInfo(screen, gm)
}

// drawHexagon draws a single hexagon with proper geometry
func drawHexagon(screen *ebiten.Image, x, y float64, hex interface{}, gm *game.GameManager) {
	size := 20.0 // Hexagon size

	// Get block color based on type if available
	blockColor := color.RGBA{100, 100, 100, 255}
	if h, ok := hex.(*world.Hexagon); ok {
		switch h.BlockType {
		case blocks.GRASS:
			blockColor = color.RGBA{34, 139, 34, 255}
		case blocks.DIRT:
			blockColor = color.RGBA{139, 90, 43, 255}
		case blocks.STONE:
			blockColor = color.RGBA{128, 128, 128, 255}
		case blocks.SAND:
			blockColor = color.RGBA{238, 203, 173, 255}
		case blocks.LOG:
			blockColor = color.RGBA{139, 69, 19, 255}
		case blocks.LEAVES:
			blockColor = color.RGBA{0, 100, 0, 255}
		case blocks.WATER:
			blockColor = color.RGBA{0, 105, 148, 200}
		case blocks.COAL_ORE:
			blockColor = color.RGBA{54, 54, 54, 255}
		case blocks.IRON_ORE:
			blockColor = color.RGBA{183, 183, 183, 255}
		case blocks.GOLD_ORE:
			blockColor = color.RGBA{255, 215, 0, 255}
		case blocks.DIAMOND_ORE:
			blockColor = color.RGBA{185, 242, 255, 255}
		}
	}

	// Draw hexagon as a filled polygon using DrawTriangles
	vertices := make([]ebiten.Vertex, 6)
	indices := []uint16{0, 1, 2, 0, 2, 3, 0, 3, 4, 0, 4, 5}

	for i := 0; i < 6; i++ {
		angle := math.Pi/3*float64(i) - math.Pi/6 // Start at top
		px := float32(x + size*math.Cos(angle))
		py := float32(y + size*math.Sin(angle))
		vertices[i] = ebiten.Vertex{
			DstX: px, DstY: py,
			SrcX: 0, SrcY: 0,
			ColorR: float32(blockColor.R) / 255,
			ColorG: float32(blockColor.G) / 255,
			ColorB: float32(blockColor.B) / 255,
			ColorA: float32(blockColor.A) / 255,
		}
	}

	// Fill hexagon
	screen.DrawTriangles(vertices, indices, gm.WhiteImage, nil)

	// Draw outline
	for i := 0; i < 6; i++ {
		angle1 := math.Pi/3*float64(i) - math.Pi/6
		angle2 := math.Pi/3*float64((i+1)%6) - math.Pi/6
		x1 := float32(x + size*math.Cos(angle1))
		y1 := float32(y + size*math.Sin(angle1))
		x2 := float32(x + size*math.Cos(angle2))
		y2 := float32(y + size*math.Sin(angle2))
		vector.StrokeLine(screen, x1, y1, x2, y2, 1, color.RGBA{50, 50, 50, 255}, true)
	}
}

// drawPlayer draws the player character
func drawPlayer(screen *ebiten.Image, x, y float64, player interface{}) {
	// Draw player as a simple character with body parts

	// Body (blue rectangle)
	ebitenutil.DrawRect(screen, x-8, y-12, 16, 20, color.RGBA{0, 100, 200, 255})

	// Head (circle approximation)
	vector.DrawFilledCircle(screen, float32(x), float32(y-18), 8, color.RGBA{255, 220, 180, 255}, true)

	// Eyes
	ebitenutil.DrawRect(screen, x-4, y-20, 3, 3, color.RGBA{0, 0, 0, 255})
	ebitenutil.DrawRect(screen, x+1, y-20, 3, 3, color.RGBA{0, 0, 0, 255})

	// Arms
	ebitenutil.DrawRect(screen, x-14, y-8, 5, 12, color.RGBA{255, 220, 180, 255})
	ebitenutil.DrawRect(screen, x+9, y-8, 5, 12, color.RGBA{255, 220, 180, 255})

	// Legs
	ebitenutil.DrawRect(screen, x-6, y+8, 5, 10, color.RGBA{100, 50, 0, 255})
	ebitenutil.DrawRect(screen, x+1, y+8, 5, 10, color.RGBA{100, 50, 0, 255})
}

// drawDroppedItem draws a dropped item with item-specific appearance
func drawDroppedItem(screen *ebiten.Image, x, y float64, item *game.DroppedItem) {
	if item == nil || item.Item == nil {
		return
	}

	// Get item color from item definitions
	itemColor := color.RGBA{200, 200, 200, 255}
	if props, ok := items.ItemDefinitions[item.Item.Type]; ok {
		itemColor = props.IconColor
	}

	// Draw item as a small square with its color
	size := 8.0
	ebitenutil.DrawRect(screen, x-size/2, y-size/2, size, size, itemColor)

	// Draw outline
	vector.StrokeRect(screen, float32(x-size/2), float32(y-size/2), float32(size), float32(size), 1, color.RGBA{50, 50, 50, 255}, true)

	// Draw quantity if more than 1
	if item.Item.Quantity > 1 {
		quantityStr := fmt.Sprintf("%d", item.Item.Quantity)
		text.Draw(screen, quantityStr, basicfont.Face7x13, int(x+2), int(y+8), color.White)
	}
}

// drawHotbar draws the inventory hotbar
func drawHotbar(screen *ebiten.Image, inventory interface{}) {
	// Draw hotbar background
	ebitenutil.DrawRect(screen, 10, float64(ScreenHeight)-60, 300, 50, color.RGBA{50, 50, 50, 200})

	// Draw hotbar slots
	for i := 0; i < 9; i++ {
		x := float64(20 + i*30)
		y := float64(ScreenHeight - 50)
		ebitenutil.DrawRect(screen, x, y, 25, 25, color.RGBA{100, 100, 100, 255})
	}
}

// drawHealthBar draws the player health bar
func drawHealthBar(screen *ebiten.Image, healthSystem interface{}) {
	// Draw health bar background
	ebitenutil.DrawRect(screen, 10, 10, 200, 20, color.RGBA{50, 50, 50, 200})

	// Draw health bar fill (placeholder - 75% health)
	ebitenutil.DrawRect(screen, 10, 10, 150, 20, color.RGBA{255, 0, 0, 255})
}

// drawDebugInfo draws debug information
func drawDebugInfo(screen *ebiten.Image, gm *game.GameManager) {
	// Draw FPS
	ebitenutil.DebugPrint(screen, "Debug Info")

	// Draw camera position
	debugText := fmt.Sprintf("Camera: %.2f, %.2f", gm.CameraX, gm.CameraY)
	text.Draw(screen, debugText, basicfont.Face7x13, 10, 50, color.White)

	// Draw music status (disabled)
	// musicStatus := "Music: "
	// if gm.BackgroundMusicManager.IsPlaying() {
	// 	musicStatus += "Playing (" + gm.BackgroundMusicManager.GetCurrentTrack() + ")"
	// } else {
	// 	musicStatus += "Stopped"
	// }
	// text.Draw(screen, musicStatus, basicfont.Face7x13, 10, 70, color.White)

	// Draw profiler info if enabled
	if gm.Profiler != nil {
		// gm.Profiler.Draw(screen)
	}
}

// Example of how to update the GameWrapper.Draw() method:
/*
func (gw *GameWrapper) Draw(screen *ebiten.Image) {
	state := gw.manager.StateManager.GetState()

	switch state {
	case ui.StateCrafting:
		// Draw game in background
		DrawGameScene(screen, gw.manager)
		// Draw crafting UI overlay
		gw.manager.CraftingUI.Draw(screen)

	case ui.StateBackpack:
		// Draw game in background
		DrawGameScene(screen, gw.manager)
		// Draw backpack UI overlay
		gw.manager.BackpackUI.Draw(screen)

	case ui.StateChest:
		// Draw game in background
		DrawGameScene(screen, gw.manager)
		// Draw chest UI overlay
		if gw.manager.ChestUI != nil {
			gw.manager.ChestUI.Draw(screen)
		}

	case ui.StatePluginUI:
		// Draw game in background
		DrawGameScene(screen, gw.manager)
		// Draw plugin UI overlay
		if gw.manager.PluginUI != nil {
			gw.manager.PluginUI.Draw(screen)
		}

	case ui.StateGame:
		// Draw main game scene
		DrawGameScene(screen, gw.manager)

		// Draw damage indicators on top
		if gw.manager.DamageIndicators != nil {
			gw.manager.DamageIndicators.Draw(screen, gw.manager.CameraX, gw.manager.CameraY)
		}

		// Draw screen flash
		if gw.manager.ScreenFlash != nil {
			gw.manager.ScreenFlash.Draw(screen, gw.screenWidth, gw.screenHeight)
		}

		// Draw directional hit indicator
		if gw.manager.DirectionalHitInd != nil {
			gw.manager.DirectionalHitInd.Draw(screen, gw.screenWidth, gw.screenHeight)
		}

		// Draw death screen on top
		if gw.manager.DeathScreen != nil {
			gw.manager.DeathScreen.Draw(screen)
		}

		// Draw profiler overlay
		gw.manager.Profiler.Draw(screen)

	case ui.StateMenu:
		// Draw menu
		// TODO: Implement menu drawing
	}
}
*/
