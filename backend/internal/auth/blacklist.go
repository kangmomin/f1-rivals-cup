package auth

import (
	"sync"
	"time"
)

// TokenBlacklist manages revoked access tokens in memory
// Tokens are stored with their expiration time and automatically cleaned up
type TokenBlacklist struct {
	mu        sync.RWMutex
	tokens    map[string]time.Time // token -> expiration time
	cleanupAt time.Time
	stopCh    chan struct{}
}

// NewTokenBlacklist creates a new in-memory token blacklist
// Starts a background cleanup goroutine that runs every 5 minutes
func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		tokens:    make(map[string]time.Time),
		cleanupAt: time.Now().Add(5 * time.Minute),
		stopCh:    make(chan struct{}),
	}
	// Start background cleanup goroutine to prevent memory leaks
	go bl.backgroundCleanup()
	return bl
}

// backgroundCleanup runs periodic cleanup to remove expired tokens
// This prevents memory leaks in long-running processes
func (bl *TokenBlacklist) backgroundCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bl.mu.Lock()
			bl.cleanup(time.Now())
			bl.mu.Unlock()
		case <-bl.stopCh:
			return
		}
	}
}

// Stop stops the background cleanup goroutine
func (bl *TokenBlacklist) Stop() {
	close(bl.stopCh)
}

// Add adds a token to the blacklist with its expiration time
// The token will be automatically removed after it expires
func (bl *TokenBlacklist) Add(token string, expiresAt time.Time) {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	// Cleanup expired tokens periodically
	now := time.Now()
	if now.After(bl.cleanupAt) {
		bl.cleanup(now)
		bl.cleanupAt = now.Add(5 * time.Minute)
	}

	bl.tokens[token] = expiresAt
}

// IsBlacklisted checks if a token is in the blacklist
func (bl *TokenBlacklist) IsBlacklisted(token string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	expiry, exists := bl.tokens[token]
	if !exists {
		return false
	}

	// Token is expired, no need to keep it in blacklist
	if time.Now().After(expiry) {
		return false
	}

	return true
}

// cleanup removes expired tokens from the blacklist
func (bl *TokenBlacklist) cleanup(now time.Time) {
	for token, expiry := range bl.tokens {
		if now.After(expiry) {
			delete(bl.tokens, token)
		}
	}
}

// Count returns the number of tokens in the blacklist (for monitoring)
func (bl *TokenBlacklist) Count() int {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return len(bl.tokens)
}
