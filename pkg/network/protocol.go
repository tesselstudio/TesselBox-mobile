package network

// PacketType represents the type of network packet
type PacketType byte

const (
	PacketTypeHandshake PacketType = iota
	PacketTypePlayerJoin
	PacketTypePlayerLeave
	PacketTypePositionUpdate
	PacketTypeBlockUpdate
	PacketTypeChatMessage
	PacketTypeWorldState
	PacketTypeError
	PacketTypeAuthChallenge
	PacketTypeAuthResponse
	PacketTypePlayerList
	PacketTypeEntityUpdate
	PacketTypeInventorySync
	PacketTypeEquipmentSync
	PacketTypeHealthSync
	PacketTypeWorldChunk
	PacketTypeTimeSync
	PacketTypeWeatherSync
	PacketTypeKick
	PacketTypeBan
	PacketTypeServerMessage
)

// Packet represents a network packet
type Packet struct {
	Type    PacketType
	Payload []byte
}

// HandshakePacket represents the initial handshake
type HandshakePacket struct {
	Version    string
	PlayerName string
	Password   string
}

// AuthChallengePacket represents server authentication challenge
type AuthChallengePacket struct {
	Challenge []byte
	Salt      []byte
}

// AuthResponsePacket represents client authentication response
type AuthResponsePacket struct {
	Response []byte
	PlayerID string
}

// PlayerJoinPacket represents a player joining
type PlayerJoinPacket struct {
	PlayerID   string
	PlayerName string
	Position   struct {
		X float64
		Y float64
	}
}

// PlayerLeavePacket represents a player leaving
type PlayerLeavePacket struct {
	PlayerID string
}

// PositionUpdatePacket represents a position update
type PositionUpdatePacket struct {
	PlayerID string
	Position struct {
		X float64
		Y float64
	}
}

// BlockUpdatePacket represents a block change
type BlockUpdatePacket struct {
	X         int
	Y         int
	BlockType string
}

// ChatMessagePacket represents a chat message
type ChatMessagePacket struct {
	PlayerID  string
	Message   string
	Timestamp int64
	MessageID string
	Channel   string // "global", "local", "private"
}

// WorldStatePacket represents world state sync
type WorldStatePacket struct {
	Seed       int64
	Difficulty string
	Players    []PlayerInfo
}

// PlayerInfo represents player information
type PlayerInfo struct {
	ID       string
	Name     string
	Position struct {
		X float64
		Y float64
	}
	Ping int
}

// PlayerListPacket represents list of connected players
type PlayerListPacket struct {
	Players []PlayerInfo
}

// EntityUpdatePacket represents entity state sync
type EntityUpdatePacket struct {
	EntityID   string
	EntityType string
	Position   struct {
		X float64
		Y float64
	}
	Velocity struct {
		X float64
		Y float64
	}
	Health float64
}

// InventorySyncPacket represents inventory sync
type InventorySyncPacket struct {
	Slots []InventorySlot
}

// InventorySlot represents an inventory slot
type InventorySlot struct {
	Index    int
	ItemType string
	Quantity int
}

// EquipmentSyncPacket represents equipment sync
type EquipmentSyncPacket struct {
	Helmet     string
	Chestplate string
	Leggings   string
	Boots      string
	Wings      string
}

// HealthSyncPacket represents health/damage sync
type HealthSyncPacket struct {
	PlayerID   string
	OverallHP  float64
	BodyParts  map[string]float64
	DamageType string
}

// WorldChunkPacket represents chunk data
type WorldChunkPacket struct {
	ChunkX    int
	ChunkY    int
	BlockData []byte
}

// TimeSyncPacket represents day/night cycle sync
type TimeSyncPacket struct {
	AmbientLight float64
	DayTime      float64
}

// WeatherSyncPacket represents weather state sync
type WeatherSyncPacket struct {
	WeatherType string
	Intensity   float64
}

// KickPacket represents server kick
type KickPacket struct {
	Reason string
}

// BanPacket represents server ban
type BanPacket struct {
	Reason   string
	Duration int64 // 0 = permanent
}

// ServerMessagePacket represents server announcement
type ServerMessagePacket struct {
	Message   string
	MessageID string
}
