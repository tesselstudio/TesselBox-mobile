package game

import (
	"encoding/json"
	"image/color"
	"log"
	"time"

	// Audio disabled: "github.com/tesselstudio/TesselBox-mobile/pkg/audio"
	"github.com/tesselstudio/TesselBox-mobile/pkg/chest"
	"github.com/tesselstudio/TesselBox-mobile/pkg/combat"
	"github.com/tesselstudio/TesselBox-mobile/pkg/crafting"
	"github.com/tesselstudio/TesselBox-mobile/pkg/debug"
	"github.com/tesselstudio/TesselBox-mobile/pkg/equipment"
	"github.com/tesselstudio/TesselBox-mobile/pkg/gametime"
	"github.com/tesselstudio/TesselBox-mobile/pkg/health"
	"github.com/tesselstudio/TesselBox-mobile/pkg/input"
	"github.com/tesselstudio/TesselBox-mobile/pkg/items"
	"github.com/tesselstudio/TesselBox-mobile/pkg/network"
	"github.com/tesselstudio/TesselBox-mobile/pkg/player"
	"github.com/tesselstudio/TesselBox-mobile/pkg/plugins"
	"github.com/tesselstudio/TesselBox-mobile/pkg/save"
	"github.com/tesselstudio/TesselBox-mobile/pkg/survival"
	"github.com/tesselstudio/TesselBox-mobile/pkg/ui"
	"github.com/tesselstudio/TesselBox-mobile/pkg/weather"
	"github.com/tesselstudio/TesselBox-mobile/pkg/world"

	"github.com/hajimehoshi/ebiten/v2"
)

// GameManager coordinates all game subsystems
type GameManager struct {
	// Core systems
	World        *world.World
	Player       *player.Player
	Inventory    *items.Inventory
	StateManager *ui.StateManager

	// Crafting
	CraftingSystem *crafting.CraftingSystem
	CraftingUI     *crafting.CraftingUI

	// Plugins
	PluginManager   *plugins.PluginManager
	PluginUI        *plugins.PluginUI
	PluginInstaller *plugins.PluginInstaller

	// Input
	InputManager *input.InputManager

	// Save system
	SaveManager *save.SaveManager
	AutoSaver   *save.AutoSaver

	// Time systems
	DayNightCycle *gametime.DayNightCycle
	WeatherSystem *weather.WeatherSystem

	// Audio (disabled)
	// AudioManager           *audio.AudioManager
	// SoundLibrary           *audio.SoundLibrary
	// BackgroundMusicManager *audio.BackgroundMusicManager

	// Debug
	RecoveryHandler *debug.RecoveryHandler
	Profiler        *debug.PerformanceProfiler

	// Survival systems
	SurvivalManager *survival.SurvivalManager
	EquipmentSet    *equipment.EquipmentSet
	HealthSystem    *health.LocationalHealthSystem
	BackpackUI      *ui.BackpackUI
	HUD             *ui.HUD

	// Chest system
	ChestManager *chest.ChestManager
	ChestUI      *ui.ChestUI

	// Combat
	WeaponSystem *combat.WeaponSystem

	// UI effects
	DamageIndicators  *ui.DamageIndicatorManager
	ScreenFlash       *ui.ScreenFlash
	DirectionalHitInd *ui.DirectionalHitManager
	DeathScreen       *ui.DeathScreen

	// Multiplayer UI
	MultiplayerConnectUI *ui.MultiplayerConnectUI

	// Enemy systems

	// Multiplayer networking
	NetworkClient *network.ClientNetwork
	IsMultiplayer bool
	RemotePlayers map[string]*RemotePlayer

	// Game state
	SelectedBlock string
	CreativeMode  bool
	CommandMode   bool
	CommandString string

	// Statistics
	BlocksPlaced    int
	BlocksDestroyed int
	ItemsCrafted    int
	PlayStartTime   time.Time
	TotalPlayTime   time.Duration

	// Crafting station tracking
	CurrentCraftingStation string
	UnlockedRecipes        map[string]bool

	// Rendering
	CameraX, CameraY float64
	CurrentLayer     int
	WhiteImage       *ebiten.Image

	// Timing
	LastTime         time.Time
	LastFootstepTime time.Time
	FootstepCooldown time.Duration
	MiningDamage     float64
	MiningSpeed      float64

	// Dropped items
	DroppedItems []*DroppedItem

	// Object pools for rendering
	VertexPool  [][]ebiten.Vertex
	IndicesPool [][]uint16
	PoolIndex   int

	// Mouse state
	MouseX, MouseY       int
	HoveredBlockName     string
	LeftMouseWasPressed  bool
	RightMouseWasPressed bool

	// Mining state
	IsMining        bool
	MiningProgress  float64
	MiningStartTime time.Time
}

// DroppedItem represents an item dropped in the world
type DroppedItem struct {
	X, Y     float64
	Item     *items.Item
	PickupAt time.Time
}

// RemotePlayer represents a player connected from another client
type RemotePlayer struct {
	ID       string
	Name     string
	Position struct {
		X float64
		Y float64
	}
}

// NewGameManager creates a new game manager with all subsystems initialized
func NewGameManager(worldName string, worldSeed int64, creativeMode bool, screenWidth, screenHeight int) *GameManager {
	log.Printf("Initializing GameManager for world '%s'...", worldName)

	gm := &GameManager{
		World:            world.NewWorld(worldName),
		Player:           player.NewPlayer(400, 300),
		Inventory:        items.NewInventory(32),
		StateManager:     ui.NewStateManager(),
		SelectedBlock:    "dirt",
		CreativeMode:     creativeMode,
		LastTime:         time.Now(),
		LastFootstepTime: time.Now(),
		FootstepCooldown: 400 * time.Millisecond,
		PlayStartTime:    time.Now(),
		UnlockedRecipes:  make(map[string]bool),
		CurrentLayer:     0,
		DroppedItems:     make([]*DroppedItem, 0),
		VertexPool:       make([][]ebiten.Vertex, 10),
		IndicesPool:      make([][]uint16, 10),
		IsMultiplayer:    false,
		RemotePlayers:    make(map[string]*RemotePlayer),
	}

	// Initialize white image for rendering
	gm.WhiteImage = ebiten.NewImage(1, 1)
	gm.WhiteImage.Fill(color.RGBA{255, 255, 255, 255})

	// Initialize object pools
	for i := range gm.VertexPool {
		gm.VertexPool[i] = make([]ebiten.Vertex, 0, 1000)
		gm.IndicesPool[i] = make([]uint16, 0, 1000)
	}

	// Set world seed
	if worldSeed != 0 {
		gm.World.SetSeed(worldSeed)
		log.Printf("World '%s' created with seed: %d", worldName, worldSeed)
	} else {
		log.Printf("World '%s' created with random seed: %d", worldName, gm.World.GetSeed())
	}

	// Find spawn position
	spawnX, spawnY := gm.World.FindSpawnPosition(0, 0)
	gm.Player.SetPosition(spawnX, spawnY)
	gm.World.GetChunksInRange(spawnX, spawnY)
	log.Printf("Player spawned at: (%.1f, %.1f)", spawnX, spawnY)

	// Initialize crafting system
	gm.CraftingSystem = crafting.NewCraftingSystem()
	if err := gm.CraftingSystem.LoadRecipesFromAssets(); err != nil {
		log.Printf("Warning: Failed to load crafting recipes: %v", err)
	}
	gm.CraftingUI = crafting.NewCraftingUI(gm.CraftingSystem, gm.Inventory)

	// Initialize input manager
	gm.InputManager = input.NewInputManager()

	// Load game assets
	items.LoadItems()

	// Add initial items
	gm.Inventory.AddItem(items.DIRT_BLOCK, 64)
	gm.Inventory.AddItem(items.STONE_BLOCK, 64)
	gm.Inventory.AddItem(items.GRASS_BLOCK, 64)
	gm.Inventory.AddItem(items.WORKBENCH, 1)
	gm.Inventory.AddItem(items.WOODEN_PICKAXE, 1)
	gm.Inventory.AddItem(items.COAL, 10)

	// Initialize plugin system
	gm.PluginManager = plugins.NewPluginManager()
	defaultPlugin := plugins.NewDefaultPlugin()
	gm.PluginManager.RegisterPlugin(defaultPlugin)
	gm.PluginManager.EnablePlugin("default")

	// Initialize save system
	gm.SaveManager = save.NewSaveManager(worldName, "player")

	// Initialize time systems
	gm.DayNightCycle = gametime.NewDayNightCycle(600.0)
	gm.WeatherSystem = weather.NewWeatherSystem()

	// Initialize audio system (disabled)
	// gm.AudioManager = audio.NewAudioManager()
	// gm.SoundLibrary = audio.NewSoundLibrary(gm.AudioManager)
	// gm.BackgroundMusicManager = audio.NewBackgroundMusicManager(gm.AudioManager)

	// Initialize debug systems
	gm.RecoveryHandler = debug.NewRecoveryHandler("", func(info debug.PanicInfo) {
		log.Printf("Recovered from panic: %s", info.Message)
	})
	gm.Profiler = debug.NewPerformanceProfiler()

	// Initialize survival systems
	gameMode := survival.ModeSurvival
	if creativeMode {
		gameMode = survival.ModeCreative
	}
	gm.SurvivalManager = survival.NewSurvivalManager(gameMode, gm.Player, gm.Inventory)
	gm.EquipmentSet = equipment.NewEquipmentSet()
	gm.HealthSystem = health.NewLocationalHealthSystem()
	gm.BackpackUI = ui.NewBackpackUI(800, 600, gm.Inventory, gm.EquipmentSet, gm.HealthSystem)

	// Initialize chest system
	gm.ChestManager = chest.NewChestManager(worldName)

	// Initialize combat
	gm.WeaponSystem = combat.NewWeaponSystem()

	// Initialize UI effects
	gm.DamageIndicators = ui.NewDamageIndicatorManager(screenWidth, screenHeight)
	gm.ScreenFlash = ui.NewScreenFlash()
	gm.DirectionalHitInd = ui.NewDirectionalHitManager()
	gm.DeathScreen = ui.NewDeathScreen(screenWidth, screenHeight)

	// Initialize multiplayer UI
	gm.MultiplayerConnectUI = ui.NewMultiplayerConnectUI()

	log.Printf("GameManager initialized successfully")
	return gm
}

// StartBackgroundMusic starts the background music loop (disabled)
func (gm *GameManager) StartBackgroundMusic() error {
	log.Printf("Audio disabled - background music not started")
	return nil
}

// StopBackgroundMusic stops the background music (disabled)
func (gm *GameManager) StopBackgroundMusic() {
	// Audio disabled
}

// Update updates all game systems
func (gm *GameManager) Update(deltaTime float64) error {
	defer gm.RecoveryHandler.Recover()

	gm.TotalPlayTime += time.Duration(deltaTime * float64(time.Second))

	state := gm.StateManager.GetState()

	switch state {
	case ui.StateCrafting:
		return gm.CraftingUI.Update()
	case ui.StateBackpack:
		return gm.BackpackUI.Update()
	case ui.StateChest:
		if gm.ChestUI != nil {
			return gm.ChestUI.Update()
		}
	case ui.StatePluginUI:
		if gm.PluginUI != nil {
			return gm.PluginUI.Update()
		}
	case ui.StateDeathScreen:
		return gm.DeathScreen.Update()
	case ui.StateMultiplayerConnect:
		if gm.MultiplayerConnectUI != nil {
			if err := gm.MultiplayerConnectUI.Update(); err != nil {
				return err
			}
			// Check if user pressed connect
			if gm.MultiplayerConnectUI.IsConnected() {
				if err := gm.ConnectToServer(
					gm.MultiplayerConnectUI.GetServerAddr(),
					gm.MultiplayerConnectUI.GetPlayerName(),
				); err != nil {
					gm.MultiplayerConnectUI.SetError(err.Error())
					gm.MultiplayerConnectUI.Reset()
				} else {
					// Connection successful, switch to game state
					gm.StateManager.SetState(ui.StateGame)
				}
			}
		}
	case ui.StateGame:
		// Update game systems
		gm.Player.Update(deltaTime)
		gm.DayNightCycle.Update()
		gm.SurvivalManager.Update(deltaTime)
		gm.HealthSystem.Update(deltaTime)
		if gm.DamageIndicators != nil {
			gm.DamageIndicators.Update(deltaTime)
		}
		if gm.ScreenFlash != nil {
			gm.ScreenFlash.Update()
		}
		if gm.DirectionalHitInd != nil {
			gm.DirectionalHitInd.Update()
		}
		// Send position updates to server if multiplayer
		if gm.IsMultiplayer {
			gm.SendPositionUpdate()
		}
	}

	// Update audio system (disabled)
	// gm.AudioManager.Update()
	// gm.BackgroundMusicManager.Update()

	return nil
}

// Cleanup cleans up all resources
func (gm *GameManager) Cleanup() {
	log.Printf("Cleaning up GameManager...")

	// Disconnect from server if connected
	if gm.NetworkClient != nil && gm.NetworkClient.IsConnected() {
		gm.NetworkClient.Disconnect()
	}

	// Audio disabled
	// gm.StopBackgroundMusic()
	// gm.AudioManager.Cleanup()
}

// ConnectToServer connects to a multiplayer server
func (gm *GameManager) ConnectToServer(serverAddr, playerName string) error {
	gm.NetworkClient = network.NewClientNetwork(serverAddr, playerName)

	if err := gm.NetworkClient.Connect(); err != nil {
		return err
	}

	gm.IsMultiplayer = true

	// Start packet listener
	gm.NetworkClient.StartPacketListener(gm.handleServerPacket)

	return nil
}

// handleServerPacket handles packets from the server
func (gm *GameManager) handleServerPacket(packet *network.Packet) {
	switch packet.Type {
	case network.PacketTypePlayerJoin:
		var joinPacket network.PlayerJoinPacket
		if err := json.Unmarshal(packet.Payload, &joinPacket); err != nil {
			log.Printf("Error unmarshaling player join: %v", err)
			return
		}
		gm.RemotePlayers[joinPacket.PlayerID] = &RemotePlayer{
			ID:   joinPacket.PlayerID,
			Name: joinPacket.PlayerName,
			Position: struct {
				X float64
				Y float64
			}{
				X: joinPacket.Position.X,
				Y: joinPacket.Position.Y,
			},
		}
		log.Printf("Remote player joined: %s (%s)", joinPacket.PlayerName, joinPacket.PlayerID)

	case network.PacketTypePlayerLeave:
		var leavePacket network.PlayerLeavePacket
		if err := json.Unmarshal(packet.Payload, &leavePacket); err != nil {
			log.Printf("Error unmarshaling player leave: %v", err)
			return
		}
		delete(gm.RemotePlayers, leavePacket.PlayerID)
		log.Printf("Remote player left: %s", leavePacket.PlayerID)

	case network.PacketTypePositionUpdate:
		var posUpdate network.PositionUpdatePacket
		if err := json.Unmarshal(packet.Payload, &posUpdate); err != nil {
			log.Printf("Error unmarshaling position update: %v", err)
			return
		}
		if player, exists := gm.RemotePlayers[posUpdate.PlayerID]; exists {
			player.Position.X = posUpdate.Position.X
			player.Position.Y = posUpdate.Position.Y
		}

	case network.PacketTypeBlockUpdate:
		var blockUpdate network.BlockUpdatePacket
		if err := json.Unmarshal(packet.Payload, &blockUpdate); err != nil {
			log.Printf("Error unmarshaling block update: %v", err)
			return
		}
		// Apply block update to world
		// TODO: Convert pixel coordinates to hex and string to BlockType
		log.Printf("Block update received: x=%d, y=%d, type=%s", blockUpdate.X, blockUpdate.Y, blockUpdate.BlockType)

	case network.PacketTypeChatMessage:
		var chatMsg network.ChatMessagePacket
		if err := json.Unmarshal(packet.Payload, &chatMsg); err != nil {
			log.Printf("Error unmarshaling chat message: %v", err)
			return
		}
		log.Printf("Chat from %s: %s", chatMsg.PlayerID, chatMsg.Message)
		// TODO: Display chat in UI

	case network.PacketTypeWorldState:
		var worldState network.WorldStatePacket
		if err := json.Unmarshal(packet.Payload, &worldState); err != nil {
			log.Printf("Error unmarshaling world state: %v", err)
			return
		}
		// Apply world state
		gm.World.SetSeed(worldState.Seed)
		log.Printf("Received world state: seed=%d, difficulty=%s", worldState.Seed, worldState.Difficulty)
	}
}

// SendPositionUpdate sends the local player's position to the server
func (gm *GameManager) SendPositionUpdate() {
	if !gm.IsMultiplayer || gm.NetworkClient == nil {
		return
	}

	x, y := gm.Player.GetPosition()
	gm.NetworkClient.SendPositionUpdate(x, y)
}

// SendBlockUpdate sends a block change to the server
func (gm *GameManager) SendBlockUpdate(x, y int, blockType string) {
	if !gm.IsMultiplayer || gm.NetworkClient == nil {
		return
	}

	gm.NetworkClient.SendBlockUpdate(x, y, blockType)
}

// SendChatMessage sends a chat message to the server
func (gm *GameManager) SendChatMessage(message string) error {
	if !gm.IsMultiplayer || gm.NetworkClient == nil {
		return nil
	}

	return gm.NetworkClient.SendChatMessage(message)
}
