package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/pkg/cache"
)

// CacheService wraps the cache package with domain-specific cache keys and TTLs
type CacheService struct {
	cache *cache.CacheService
}

// NewCacheService creates a new cache service
func NewCacheService() *CacheService {
	return &CacheService{
		cache: cache.NewCacheService(),
	}
}

// Cache key prefixes
const (
	CacheKeyDashboardStats = "dashboard:stats"
	CacheKeyCalendarMonth  = "calendar:month"
	CacheKeyProjectStats   = "project:stats"
	CacheKeyActivityLogs   = "activity:logs"
)

// Cache TTLs
const (
	TTLDashboardStats = 3 * time.Minute
	TTLCalendarMonth  = 10 * time.Minute
	TTLProjectStats   = 2 * time.Minute
	TTLActivityLogs   = 1 * time.Minute
)

// GetDashboardStats retrieves dashboard stats from cache
func (s *CacheService) GetDashboardStats(ctx context.Context, projectLimit, empLimit int) (map[string]interface{}, error) {
	key := fmt.Sprintf("%s:project_limit:%d:emp_limit:%d", CacheKeyDashboardStats, projectLimit, empLimit)
	var result map[string]interface{}
	err := s.cache.Get(ctx, key, &result)
	if err == cache.ErrCacheMiss {
		return nil, nil
	}
	return result, err
}

// SetDashboardStats stores dashboard stats in cache
func (s *CacheService) SetDashboardStats(ctx context.Context, projectLimit, empLimit int, stats map[string]interface{}) error {
	key := fmt.Sprintf("%s:project_limit:%d:emp_limit:%d", CacheKeyDashboardStats, projectLimit, empLimit)
	return s.cache.Set(ctx, key, stats, TTLDashboardStats)
}

// InvalidateDashboardStats invalidates dashboard stats cache
func (s *CacheService) InvalidateDashboardStats(ctx context.Context) error {
	pattern := fmt.Sprintf("%s:*", CacheKeyDashboardStats)
	return s.cache.DeletePattern(ctx, pattern)
}

// GetCalendarMonth retrieves calendar events for a month from cache
func (s *CacheService) GetCalendarMonth(ctx context.Context, year, month int) ([]interface{}, error) {
	key := fmt.Sprintf("%s:%d:%d", CacheKeyCalendarMonth, year, month)
	var result []interface{}
	err := s.cache.Get(ctx, key, &result)
	if err == cache.ErrCacheMiss {
		return nil, nil
	}
	return result, err
}

// SetCalendarMonth stores calendar events for a month in cache
func (s *CacheService) SetCalendarMonth(ctx context.Context, year, month int, events []interface{}) error {
	key := fmt.Sprintf("%s:%d:%d", CacheKeyCalendarMonth, year, month)
	// Use longer TTL for past months
	ttl := TTLCalendarMonth
	now := time.Now()
	eventMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	if eventMonth.Before(now.AddDate(0, -1, 0)) {
		// Past months: cache for 1 hour
		ttl = 1 * time.Hour
	}
	return s.cache.Set(ctx, key, events, ttl)
}

// InvalidateCalendarMonth invalidates calendar events cache for a specific month
func (s *CacheService) InvalidateCalendarMonth(ctx context.Context, year, month int) error {
	key := fmt.Sprintf("%s:%d:%d", CacheKeyCalendarMonth, year, month)
	return s.cache.Delete(ctx, key)
}

// InvalidateAllCalendarEvents invalidates all calendar events cache
func (s *CacheService) InvalidateAllCalendarEvents(ctx context.Context) error {
	pattern := fmt.Sprintf("%s:*", CacheKeyCalendarMonth)
	return s.cache.DeletePattern(ctx, pattern)
}

// GetProjectStats retrieves project statistics from cache
func (s *CacheService) GetProjectStats(ctx context.Context) ([]interface{}, error) {
	key := CacheKeyProjectStats
	var result []interface{}
	err := s.cache.Get(ctx, key, &result)
	if err == cache.ErrCacheMiss {
		return nil, nil
	}
	return result, err
}

// SetProjectStats stores project statistics in cache
func (s *CacheService) SetProjectStats(ctx context.Context, stats []interface{}) error {
	key := CacheKeyProjectStats
	return s.cache.Set(ctx, key, stats, TTLProjectStats)
}

// InvalidateProjectStats invalidates project statistics cache
func (s *CacheService) InvalidateProjectStats(ctx context.Context) error {
	return s.cache.Delete(ctx, CacheKeyProjectStats)
}

// GetActivityLogs retrieves activity logs from cache
func (s *CacheService) GetActivityLogs(ctx context.Context, entityType string, limit int) ([]interface{}, error) {
	key := fmt.Sprintf("%s:%s:limit:%d", CacheKeyActivityLogs, entityType, limit)
	var result []interface{}
	err := s.cache.Get(ctx, key, &result)
	if err == cache.ErrCacheMiss {
		return nil, nil
	}
	return result, err
}

// SetActivityLogs stores activity logs in cache
func (s *CacheService) SetActivityLogs(ctx context.Context, entityType string, limit int, logs []interface{}) error {
	key := fmt.Sprintf("%s:%s:limit:%d", CacheKeyActivityLogs, entityType, limit)
	return s.cache.Set(ctx, key, logs, TTLActivityLogs)
}

// InvalidateActivityLogs invalidates activity logs cache
func (s *CacheService) InvalidateActivityLogs(ctx context.Context, entityType string) error {
	pattern := fmt.Sprintf("%s:%s:*", CacheKeyActivityLogs, entityType)
	return s.cache.DeletePattern(ctx, pattern)
}

// InvalidateAllActivityLogs invalidates all activity logs cache
func (s *CacheService) InvalidateAllActivityLogs(ctx context.Context) error {
	pattern := fmt.Sprintf("%s:*", CacheKeyActivityLogs)
	return s.cache.DeletePattern(ctx, pattern)
}

// Item-level caching for frequently accessed entities

// Cache key prefixes for individual items
const (
	CacheKeyUser    = "user"
	CacheKeyProject = "project"
	CacheKeyTask    = "task"
)

// Cache TTLs for individual items
const (
	TTLUser    = 10 * time.Minute
	TTLProject = 10 * time.Minute
	TTLTask    = 5 * time.Minute
)

// GetUser retrieves a user from cache
func (s *CacheService) GetUser(ctx context.Context, userID string, dest interface{}) error {
	key := fmt.Sprintf("%s:%s", CacheKeyUser, userID)
	err := s.cache.Get(ctx, key, dest)
	if err == cache.ErrCacheMiss {
		return cache.ErrCacheMiss
	}
	return err
}

// SetUser stores a user in cache
func (s *CacheService) SetUser(ctx context.Context, userID string, user interface{}) error {
	key := fmt.Sprintf("%s:%s", CacheKeyUser, userID)
	return s.cache.Set(ctx, key, user, TTLUser)
}

// InvalidateUser invalidates a user cache
func (s *CacheService) InvalidateUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s:%s", CacheKeyUser, userID)
	return s.cache.Delete(ctx, key)
}

// GetProject retrieves a project from cache
func (s *CacheService) GetProject(ctx context.Context, projectID string, dest interface{}) error {
	key := fmt.Sprintf("%s:%s", CacheKeyProject, projectID)
	err := s.cache.Get(ctx, key, dest)
	if err == cache.ErrCacheMiss {
		return cache.ErrCacheMiss
	}
	return err
}

// SetProject stores a project in cache
func (s *CacheService) SetProject(ctx context.Context, projectID string, project interface{}) error {
	key := fmt.Sprintf("%s:%s", CacheKeyProject, projectID)
	return s.cache.Set(ctx, key, project, TTLProject)
}

// InvalidateProject invalidates a project cache
func (s *CacheService) InvalidateProject(ctx context.Context, projectID string) error {
	key := fmt.Sprintf("%s:%s", CacheKeyProject, projectID)
	return s.cache.Delete(ctx, key)
}

// GetTask retrieves a task from cache
func (s *CacheService) GetTask(ctx context.Context, taskID string, dest interface{}) error {
	key := fmt.Sprintf("%s:%s", CacheKeyTask, taskID)
	err := s.cache.Get(ctx, key, dest)
	if err == cache.ErrCacheMiss {
		return cache.ErrCacheMiss
	}
	return err
}

// SetTask stores a task in cache
func (s *CacheService) SetTask(ctx context.Context, taskID string, task interface{}) error {
	key := fmt.Sprintf("%s:%s", CacheKeyTask, taskID)
	return s.cache.Set(ctx, key, task, TTLTask)
}

// InvalidateTask invalidates a task cache
func (s *CacheService) InvalidateTask(ctx context.Context, taskID string) error {
	key := fmt.Sprintf("%s:%s", CacheKeyTask, taskID)
	return s.cache.Delete(ctx, key)
}

// FlushAll clears all cache entries from Redis
func (s *CacheService) FlushAll(ctx context.Context) error {
	return s.cache.FlushAll(ctx)
}