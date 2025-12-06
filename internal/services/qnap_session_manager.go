package services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// QNAPSessionManager manages QNAP sessions with caching and auto-refresh
type QNAPSessionManager struct {
	client           *QNAPClient
	mu               sync.RWMutex
	sid              string
	expiresAt        time.Time
	refreshThreshold time.Duration // Refresh session this much before expiration
}

// NewQNAPSessionManager creates a new QNAP session manager
func NewQNAPSessionManager(client *QNAPClient) *QNAPSessionManager {
	return &QNAPSessionManager{
		client:           client,
		refreshThreshold: 5 * time.Minute, // Refresh 5 minutes before expiration
	}
}

// GetSessionID returns a valid session ID, refreshing if necessary
func (m *QNAPSessionManager) GetSessionID(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we have a valid session
	now := time.Now()
	if m.sid != "" && now.Before(m.expiresAt) {
		// Check if we need to refresh soon
		if now.Add(m.refreshThreshold).After(m.expiresAt) {
			// Refresh in background (don't block)
			go m.refreshSession()
		}
		return m.sid, nil
	}

	// Need to authenticate
	if err := m.client.Authenticate(); err != nil {
		return "", fmt.Errorf("failed to authenticate with QNAP: %w", err)
	}

	m.sid = m.client.sid
	m.expiresAt = m.client.expiresAt

	return m.sid, nil
}

// refreshSession refreshes the session in the background
func (m *QNAPSessionManager) refreshSession() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check we still need to refresh
	if time.Now().Add(m.refreshThreshold).Before(m.expiresAt) {
		return
	}

	if err := m.client.RefreshSession(); err != nil {
		// Log error but don't fail - will retry on next request
		fmt.Printf("Warning: Failed to refresh QNAP session: %v\n", err)
		return
	}

	m.sid = m.client.sid
	m.expiresAt = m.client.expiresAt
}

// InvalidateSession invalidates the current session
func (m *QNAPSessionManager) InvalidateSession() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sid = ""
	m.expiresAt = time.Time{}
}

// IsSessionValid checks if the current session is valid
func (m *QNAPSessionManager) IsSessionValid() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sid != "" && time.Now().Before(m.expiresAt)
}
