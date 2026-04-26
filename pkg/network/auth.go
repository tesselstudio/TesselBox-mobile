package network

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// AuthManager handles authentication for both client and server
type AuthManager struct {
	serverSecret []byte // Server's secret key for signing challenges
	sessions    map[string]*Session
}

// Session represents an authenticated session
type Session struct {
	PlayerID    string
	PlayerName string
	Token       string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	secret := make([]byte, 32)
	rand.Read(secret)
	
	return &AuthManager{
		serverSecret: secret,
		sessions:    make(map[string]*Session),
	}
}

// GenerateChallenge generates a random challenge for authentication
func (am *AuthManager) GenerateChallenge() ([]byte, []byte, error) {
	challenge := make([]byte, 32)
	salt := make([]byte, 16)
	
	if _, err := rand.Read(challenge); err != nil {
		return nil, nil, fmt.Errorf("failed to generate challenge: %w", err)
	}
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	
	return challenge, salt, nil
}

// DerivePlayerSecret derives a player-specific secret from password and salt
func DerivePlayerSecret(password string, salt []byte) []byte {
	h := hmac.New(sha256.New, []byte(password))
	h.Write(salt)
	return h.Sum(nil)
}

// SignChallenge signs a challenge with the player's secret
func SignChallenge(challenge []byte, playerSecret []byte) []byte {
	h := hmac.New(sha256.New, playerSecret)
	h.Write(challenge)
	return h.Sum(nil)
}

// VerifyChallenge verifies a client's response to a challenge
func (am *AuthManager) VerifyChallenge(challenge, response []byte, playerSecret []byte) bool {
	expected := SignChallenge(challenge, playerSecret)
	return hmac.Equal(response, expected)
}

// CreateSession creates a new session for an authenticated player
func (am *AuthManager) CreateSession(playerID, playerName string) (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	tokenStr := hex.EncodeToString(token)
	
	session := &Session{
		PlayerID:    playerID,
		PlayerName:  playerName,
		Token:       tokenStr,
		ExpiresAt:   time.Now().Add(24 * time.Hour), // 24 hour session
		CreatedAt:   time.Now(),
	}
	
	am.sessions[playerID] = session
	return tokenStr, nil
}

// ValidateSession validates a session token
func (am *AuthManager) ValidateSession(playerID, token string) bool {
	session, exists := am.sessions[playerID]
	if !exists {
		return false
	}
	
	if time.Now().After(session.ExpiresAt) {
		delete(am.sessions, playerID)
		return false
	}
	
	return session.Token == token
}

// RefreshSession refreshes a session's expiration
func (am *AuthManager) RefreshSession(playerID string) error {
	session, exists := am.sessions[playerID]
	if !exists {
		return fmt.Errorf("session not found")
	}
	
	session.ExpiresAt = time.Now().Add(24 * time.Hour)
	return nil
}

// RevokeSession revokes a player's session
func (am *AuthManager) RevokeSession(playerID string) {
	delete(am.sessions, playerID)
}

// CleanupExpiredSessions removes expired sessions
func (am *AuthManager) CleanupExpiredSessions() {
	now := time.Now()
	for playerID, session := range am.sessions {
		if now.After(session.ExpiresAt) {
			delete(am.sessions, playerID)
		}
	}
}

// GetSessionCount returns the number of active sessions
func (am *AuthManager) GetSessionCount() int {
	return len(am.sessions)
}
