package network

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// ClientNetwork handles client-side networking
type ClientNetwork struct {
	serverAddr string
	conn       net.Conn
	playerID   string
	playerName string
}

// NewClientNetwork creates a new client network instance
func NewClientNetwork(serverAddr, playerName string) *ClientNetwork {
	return &ClientNetwork{
		serverAddr: serverAddr,
		playerName: playerName,
	}
}

// Connect connects to the server
func (cn *ClientNetwork) Connect() error {
	conn, err := net.Dial("tcp", cn.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	cn.conn = conn
	log.Printf("Connected to server at %s", cn.serverAddr)

	// Send handshake
	handshake := HandshakePacket{
		Version:    "1.0",
		PlayerName: cn.playerName,
	}

	payload, err := json.Marshal(handshake)
	if err != nil {
		return fmt.Errorf("failed to marshal handshake: %w", err)
	}

	packet := &Packet{
		Type:    PacketTypeHandshake,
		Payload: payload,
	}

	if err := cn.writePacket(packet); err != nil {
		return fmt.Errorf("failed to send handshake: %w", err)
	}

	// Receive world state
	// TODO: Receive and process world state

	return nil
}

// Disconnect disconnects from the server
func (cn *ClientNetwork) Disconnect() error {
	if cn.conn != nil {
		return cn.conn.Close()
	}
	return nil
}

// SendPositionUpdate sends a position update to the server
func (cn *ClientNetwork) SendPositionUpdate(x, y float64) error {
	posUpdate := PositionUpdatePacket{
		PlayerID: cn.playerID,
	}
	posUpdate.Position.X = x
	posUpdate.Position.Y = y

	payload, err := json.Marshal(posUpdate)
	if err != nil {
		return err
	}

	packet := &Packet{
		Type:    PacketTypePositionUpdate,
		Payload: payload,
	}

	return cn.writePacket(packet)
}

// SendBlockUpdate sends a block update to the server
func (cn *ClientNetwork) SendBlockUpdate(x, y int, blockType string) error {
	blockUpdate := BlockUpdatePacket{
		X:         x,
		Y:         y,
		BlockType: blockType,
	}

	payload, err := json.Marshal(blockUpdate)
	if err != nil {
		return err
	}

	packet := &Packet{
		Type:    PacketTypeBlockUpdate,
		Payload: payload,
	}

	return cn.writePacket(packet)
}

// SendChatMessage sends a chat message to the server
func (cn *ClientNetwork) SendChatMessage(message string) error {
	chatMsg := ChatMessagePacket{
		PlayerID: cn.playerID,
		Message:  message,
	}

	payload, err := json.Marshal(chatMsg)
	if err != nil {
		return err
	}

	packet := &Packet{
		Type:    PacketTypeChatMessage,
		Payload: payload,
	}

	return cn.writePacket(packet)
}

// readPacket reads a packet from the server
func (cn *ClientNetwork) readPacket() (*Packet, error) {
	// Read packet type
	typeBuf := make([]byte, 1)
	if _, err := cn.conn.Read(typeBuf); err != nil {
		return nil, err
	}

	// Read payload length
	lengthBuf := make([]byte, 4)
	if _, err := cn.conn.Read(lengthBuf); err != nil {
		return nil, err
	}

	payloadLength := binary.BigEndian.Uint32(lengthBuf)

	// Read payload
	payload := make([]byte, payloadLength)
	if _, err := cn.conn.Read(payload); err != nil {
		return nil, err
	}

	return &Packet{
		Type:    PacketType(typeBuf[0]),
		Payload: payload,
	}, nil
}

// writePacket writes a packet to the server
func (cn *ClientNetwork) writePacket(packet *Packet) error {
	if cn.conn == nil {
		return fmt.Errorf("not connected to server")
	}

	// Write packet type
	if _, err := cn.conn.Write([]byte{byte(packet.Type)}); err != nil {
		return err
	}

	// Write payload length
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(packet.Payload)))
	if _, err := cn.conn.Write(lengthBuf); err != nil {
		return err
	}

	// Write payload
	if _, err := cn.conn.Write(packet.Payload); err != nil {
		return err
	}

	return nil
}

// StartPacketListener starts listening for packets from the server
func (cn *ClientNetwork) StartPacketListener(packetHandler func(*Packet)) {
	go func() {
		for {
			packet, err := cn.readPacket()
			if err != nil {
				log.Printf("Error reading packet from server: %v", err)
				return
			}
			packetHandler(packet)
		}
	}()
}

// IsConnected returns whether the client is connected
func (cn *ClientNetwork) IsConnected() bool {
	return cn.conn != nil
}
