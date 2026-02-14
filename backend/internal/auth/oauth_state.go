package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/google/uuid"
)

// OAuthStateEntry represents a pending OAuth state entry
type OAuthStateEntry struct {
	Purpose   string     // "login" or "link"
	UserID    *uuid.UUID // non-nil when purpose is "link"
	ExpiresAt time.Time
}

// OAuthState manages OAuth state tokens in memory
// Follows the same pattern as TokenBlacklist
type OAuthState struct {
	mu     sync.RWMutex
	states map[string]OAuthStateEntry
	stopCh chan struct{}
}

// NewOAuthState creates a new in-memory OAuth state manager
// Starts a background cleanup goroutine that runs every 5 minutes
func NewOAuthState() *OAuthState {
	s := &OAuthState{
		states: make(map[string]OAuthStateEntry),
		stopCh: make(chan struct{}),
	}
	go s.backgroundCleanup()
	return s
}

// backgroundCleanup runs periodic cleanup to remove expired state entries
func (s *OAuthState) backgroundCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			s.cleanup(time.Now())
			s.mu.Unlock()
		case <-s.stopCh:
			return
		}
	}
}

// Stop stops the background cleanup goroutine
func (s *OAuthState) Stop() {
	close(s.stopCh)
}

// Generate creates a new state token with the given purpose and optional user ID
func (s *OAuthState) Generate(purpose string, userID *uuid.UUID) string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to UUID if crypto/rand fails
		return uuid.New().String()
	}
	state := hex.EncodeToString(bytes)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.states[state] = OAuthStateEntry{
		Purpose:   purpose,
		UserID:    userID,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	return state
}

// Validate validates and consumes a state token (one-time use)
func (s *OAuthState) Validate(state string) (*OAuthStateEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.states[state]
	if !exists {
		return nil, false
	}

	// Delete after consumption (one-time use)
	delete(s.states, state)

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return &entry, true
}

// cleanup removes expired state entries
func (s *OAuthState) cleanup(now time.Time) {
	for state, entry := range s.states {
		if now.After(entry.ExpiresAt) {
			delete(s.states, state)
		}
	}
}
