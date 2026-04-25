package ui

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
)

// MultiplayerConnectUI handles the multiplayer server connection screen
type MultiplayerConnectUI struct {
	serverAddr string
	playerName string
	connected  bool
	errorMsg   string
	font       font.Face
}

// NewMultiplayerConnectUI creates a new multiplayer connection UI
func NewMultiplayerConnectUI() *MultiplayerConnectUI {
	// Load font
	fontFace, err := loadDefaultFont()
	if err != nil {
		log.Printf("Warning: Failed to load font: %v", err)
	}

	return &MultiplayerConnectUI{
		serverAddr: "localhost:25565",
		playerName: "Player",
		font:       fontFace,
	}
}

// Update handles input updates
func (mc *MultiplayerConnectUI) Update() error {
	// Handle keyboard input for server address
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if len(mc.serverAddr) > 0 {
			mc.serverAddr = mc.serverAddr[:len(mc.serverAddr)-1]
		}
	}

	// Handle keyboard input for player name
	// TODO: Implement proper text input handling

	// Handle Enter key to connect
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		mc.connected = true
		mc.errorMsg = ""
	}

	// Handle Escape key to cancel
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		mc.connected = false
	}

	return nil
}

// Draw renders the multiplayer connection screen
func (mc *MultiplayerConnectUI) Draw(screen *ebiten.Image) {
	// TODO: Implement drawing of connection UI
	// This should show:
	// - Server address input field
	// - Player name input field
	// - Connect button
	// - Cancel button
	// - Error message if connection failed
}

// GetServerAddr returns the entered server address
func (mc *MultiplayerConnectUI) GetServerAddr() string {
	return mc.serverAddr
}

// GetPlayerName returns the entered player name
func (mc *MultiplayerConnectUI) GetPlayerName() string {
	return mc.playerName
}

// IsConnected returns whether the user pressed connect
func (mc *MultiplayerConnectUI) IsConnected() bool {
	return mc.connected
}

// SetError sets an error message to display
func (mc *MultiplayerConnectUI) SetError(msg string) {
	mc.errorMsg = msg
}

// Reset resets the UI state
func (mc *MultiplayerConnectUI) Reset() {
	mc.connected = false
	mc.errorMsg = ""
}

// loadDefaultFont loads a default font for the UI
func loadDefaultFont() (font.Face, error) {
	// TODO: Load an actual font file
	// For now, return nil - the drawing code should handle this
	return nil, fmt.Errorf("font loading not implemented")
}
