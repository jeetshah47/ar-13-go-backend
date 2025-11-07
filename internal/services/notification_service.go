package services

import (
	"context"
	"sync"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
)

// countCacheEntry represents a cached count entry
type countCacheEntry struct {
	count     map[string]int
	timestamp time.Time
}

// NotificationService handles notification business logic
type NotificationService struct {
	notificationRepo *repos.NotificationRepo
	countCache       map[string]countCacheEntry
	cacheMutex       sync.RWMutex
	cacheTTL         time.Duration
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{
		notificationRepo: repos.NewNotificationRepo(),
		countCache:       make(map[string]countCacheEntry),
		cacheTTL:         30 * time.Second, // Cache for 30 seconds
	}
}

// invalidateCountCache invalidates the count cache for a user
func (s *NotificationService) invalidateCountCache(userID string) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	delete(s.countCache, userID)
}

// GetAll gets all notifications for a user
func (s *NotificationService) GetAll(ctx context.Context, userID string) ([]models.Notification, error) {
	return s.notificationRepo.GetAll(ctx, userID)
}

// GetUnread gets unread notifications for a user
func (s *NotificationService) GetUnread(ctx context.Context, userID string) ([]models.Notification, error) {
	return s.notificationRepo.GetUnread(ctx, userID)
}

// GetCount gets notification count for a user
// Uses caching to prevent quota exhaustion - caches results for 30 seconds
func (s *NotificationService) GetCount(ctx context.Context, userID string) (map[string]int, error) {
	// Check cache first
	s.cacheMutex.RLock()
	if entry, exists := s.countCache[userID]; exists {
		if time.Since(entry.timestamp) < s.cacheTTL {
			s.cacheMutex.RUnlock()
			return entry.count, nil
		}
	}
	s.cacheMutex.RUnlock()

	// Cache miss or expired - fetch from database
	total, unread, err := s.notificationRepo.GetCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	count := map[string]int{
		"total":  total,
		"unread": unread,
	}

	// Update cache
	s.cacheMutex.Lock()
	s.countCache[userID] = countCacheEntry{
		count:     count,
		timestamp: time.Now(),
	}
	s.cacheMutex.Unlock()

	return count, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, id string) error {
	// Get notification to find userID for cache invalidation
	notification, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if notification == nil {
		return nil
	}

	err = s.notificationRepo.MarkAsRead(ctx, id)
	if err == nil {
		// Invalidate cache for this user
		s.invalidateCountCache(notification.UserID)
	}
	return err
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	err := s.notificationRepo.MarkAllAsRead(ctx, userID)
	if err == nil {
		// Invalidate cache for this user
		s.invalidateCountCache(userID)
	}
	return err
}

// Delete deletes a notification
func (s *NotificationService) Delete(ctx context.Context, id string) error {
	// Get notification to find userID for cache invalidation
	notification, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if notification == nil {
		return nil
	}

	err = s.notificationRepo.Delete(ctx, id)
	if err == nil {
		// Invalidate cache for this user
		s.invalidateCountCache(notification.UserID)
	}
	return err
}

// DeleteAllForUser deletes all notifications for a user
func (s *NotificationService) DeleteAllForUser(ctx context.Context, userID string) error {
	err := s.notificationRepo.DeleteAllForUser(ctx, userID)
	if err == nil {
		// Invalidate cache for this user
		s.invalidateCountCache(userID)
	}
	return err
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, notification *models.Notification) error {
	err := s.notificationRepo.Add(ctx, notification)
	if err == nil {
		// Invalidate cache for this user
		s.invalidateCountCache(notification.UserID)
	}
	return err
}

// CreateUserLoginNotification creates a login notification
func (s *NotificationService) CreateUserLoginNotification(ctx context.Context, userID, userEmail string) error {
	notification := &models.Notification{
		Title:             "Login Successful",
		Message:           "You have successfully logged in",
		Type:              models.NotificationTypeUserLogin,
		UserID:            userID,
		RelatedEntityID:   userID,
		RelatedEntityType: models.RelatedEntityTypeUser,
		IsRead:            false,
	}
	err := s.notificationRepo.Add(ctx, notification)
	if err == nil {
		// Invalidate cache for this user
		s.invalidateCountCache(userID)
	}
	return err
}

// CreateUserLogoutNotification creates a logout notification
func (s *NotificationService) CreateUserLogoutNotification(ctx context.Context, userID, userEmail string) error {
	notification := &models.Notification{
		Title:             "Logout Successful",
		Message:           "You have successfully logged out",
		Type:              models.NotificationTypeUserLogout,
		UserID:            userID,
		RelatedEntityID:   userID,
		RelatedEntityType: models.RelatedEntityTypeUser,
		IsRead:            false,
	}
	err := s.notificationRepo.Add(ctx, notification)
	if err == nil {
		// Invalidate cache for this user
		s.invalidateCountCache(userID)
	}
	return err
}
