package network

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
)

// EncryptionManager handles encryption for network connections
type EncryptionManager struct {
	privateKey   *ecdh.PrivateKey
	peerKey      *ecdh.PublicKey
	sharedSecret []byte
	cipher       cipher.AEAD
	initialized  bool
}

// NewEncryptionManager creates a new encryption manager
func NewEncryptionManager() (*EncryptionManager, error) {
	privateKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	
	return &EncryptionManager{
		privateKey:  privateKey,
		initialized: false,
	}, nil
}

// GetPublicKey returns the public key for key exchange
func (em *EncryptionManager) GetPublicKey() ([]byte, error) {
	return em.privateKey.PublicKey().Bytes(), nil
}

// DeriveSharedSecret derives a shared secret from peer's public key
func (em *EncryptionManager) DeriveSharedSecret(peerPublicKey []byte) error {
	peerKey, err := ecdh.P256().NewPublicKey(peerPublicKey)
	if err != nil {
		return fmt.Errorf("invalid peer public key: %w", err)
	}
	
	em.peerKey = peerKey
	sharedSecret, err := em.privateKey.ECDH(peerKey)
	if err != nil {
		return fmt.Errorf("ECDH failed: %w", err)
	}
	
	// Derive encryption key from shared secret using HKDF-like approach
	h := sha256.New()
	h.Write(sharedSecret)
	em.sharedSecret = h.Sum(nil)
	
	// Initialize AES-GCM cipher
	block, err := aes.NewCipher(em.sharedSecret)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}
	
	em.cipher, err = cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}
	
	em.initialized = true
	return nil
}

// Encrypt encrypts data using AES-256-GCM
func (em *EncryptionManager) Encrypt(plaintext []byte) ([]byte, error) {
	if !em.initialized {
		return nil, fmt.Errorf("encryption not initialized")
	}
	
	nonce := make([]byte, em.cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	
	ciphertext := em.cipher.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-256-GCM
func (em *EncryptionManager) Decrypt(ciphertext []byte) ([]byte, error) {
	if !em.initialized {
		return nil, fmt.Errorf("encryption not initialized")
	}
	
	nonceSize := em.cipher.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := em.cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}
	
	return plaintext, nil
}

// IsInitialized returns whether encryption is initialized
func (em *EncryptionManager) IsInitialized() bool {
	return em.initialized
}

// EncryptPacket encrypts a packet with header
func (em *EncryptionManager) EncryptPacket(packetType byte, payload []byte) ([]byte, error) {
	// Format: [packetType (1 byte)] [payload length (4 bytes)] [encrypted payload]
	header := make([]byte, 5)
	header[0] = packetType
	binary.BigEndian.PutUint32(header[1:5], uint32(len(payload)))
	
	encryptedPayload, err := em.Encrypt(payload)
	if err != nil {
		return nil, err
	}
	
	result := append(header, encryptedPayload...)
	return result, nil
}

// DecryptPacket decrypts a packet with header
func (em *EncryptionManager) DecryptPacket(data []byte) (byte, []byte, error) {
	if len(data) < 5 {
		return 0, nil, fmt.Errorf("packet too short")
	}
	
	packetType := data[0]
	payloadLength := binary.BigEndian.Uint32(data[1:5])
	encryptedPayload := data[5:]
	
	payload, err := em.Decrypt(encryptedPayload)
	if err != nil {
		return 0, nil, err
	}
	
	if uint32(len(payload)) != payloadLength {
		return 0, nil, fmt.Errorf("payload length mismatch")
	}
	
	return packetType, payload, nil
}
