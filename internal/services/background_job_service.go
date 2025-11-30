package services

import (
	"context"
	"log"
	"time"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/repos"
)

// BackgroundJobService handles periodic background jobs
type BackgroundJobService struct {
	timeTrackingService *TimeTrackingService
	taskRepo            repos.TaskRepository
	interval            time.Duration // Default: 5 minutes
	idleTimeout         time.Duration // Default: 10 minutes
	stopChan            chan bool
	running             bool
}

// NewBackgroundJobService creates a new background job service
func NewBackgroundJobService(
	timeTrackingService *TimeTrackingService,
	taskRepo repos.TaskRepository,
) *BackgroundJobService {
	return &BackgroundJobService{
		timeTrackingService: timeTrackingService,
		taskRepo:            taskRepo,
		interval:            5 * time.Minute, // Default 5 minutes
		idleTimeout:         10 * time.Minute, // Default 10 minutes
		stopChan:             make(chan bool),
		running:              false,
	}
}

// Start starts the background job service
func (s *BackgroundJobService) Start(ctx context.Context) {
	if s.running {
		log.Println("Background job service is already running")
		return
	}

	s.running = true
	log.Println("Starting background job service for time tracking aggregation")

	go s.run(ctx)
}

// Stop stops the background job service
func (s *BackgroundJobService) Stop() {
	if !s.running {
		return
	}

	log.Println("Stopping background job service")
	s.stopChan <- true
	s.running = false
}

// run runs the background job loop
func (s *BackgroundJobService) run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run immediately on start
	s.processTimeTracking(ctx)

	for {
		select {
		case <-ticker.C:
			s.processTimeTracking(ctx)
		case <-s.stopChan:
			log.Println("Background job service stopped")
			return
		case <-ctx.Done():
			log.Println("Background job service context cancelled")
			return
		}
	}
}

// processTimeTracking processes all active time tracking sessions
func (s *BackgroundJobService) processTimeTracking(ctx context.Context) {
	log.Println("Processing time tracking sessions...")

	// Get all active sessions
	sessions, err := s.timeTrackingService.GetAllActiveSessions(ctx)
	if err != nil {
		log.Printf("Error getting active sessions: %v", err)
		return
	}

	if len(sessions) == 0 {
		log.Println("No active time tracking sessions found")
		return
	}

	log.Printf("Found %d active time tracking sessions", len(sessions))

	// Group sessions by task
	taskSessions := make(map[string][]string) // taskID -> []sessionID
	for _, session := range sessions {
		key := session.ProjectID + ":" + session.TaskID
		taskSessions[key] = append(taskSessions[key], session.ID)
	}

		// Process each task
	for key := range taskSessions {
		// Extract projectID and taskID from key
		// Format: "projectID:taskID"
		var projectID, taskID string
		for i, char := range key {
			if char == ':' {
				projectID = key[:i]
				taskID = key[i+1:]
				break
			}
		}

		// Verify task is still in progress
		task, err := s.taskRepo.GetByID(ctx, projectID, taskID)
		if err != nil {
			log.Printf("Error getting task %s: %v", taskID, err)
			continue
		}
		if task == nil {
			log.Printf("Task %s not found, stopping sessions", taskID)
			// Stop all sessions for this task
			activeSessions, _ := s.timeTrackingService.GetAllActiveSessions(ctx)
			for _, session := range activeSessions {
				if session.ProjectID == projectID && session.TaskID == taskID {
					_ = s.timeTrackingService.StopTracking(ctx, projectID, taskID, session.UserID)
				}
			}
			continue
		}

		// If task is no longer in progress, stop tracking
		if task.Status != string(constants.TaskStatusInProgress) {
			log.Printf("Task %s status is %s, stopping sessions", taskID, task.Status)
			// Get all active sessions for this task and stop them
			activeSessions, _ := s.timeTrackingService.GetAllActiveSessions(ctx)
			for _, session := range activeSessions {
				if session.ProjectID == projectID && session.TaskID == taskID {
					_ = s.timeTrackingService.StopTracking(ctx, projectID, taskID, session.UserID)
				}
			}
			continue
		}

		// Aggregate time for this task
		if err := s.timeTrackingService.AggregateTime(ctx, projectID, taskID); err != nil {
			log.Printf("Error aggregating time for task %s: %v", taskID, err)
			continue
		}

		// Check for idle sessions and handle them
		activeSessions, _ := s.timeTrackingService.GetAllActiveSessions(ctx)
		for _, session := range activeSessions {
			if session.ProjectID == projectID && session.TaskID == taskID {
				// Check if session is idle
				now := time.Now()
				timeSinceLastActive := now.Sub(session.LastActive)

				if timeSinceLastActive > s.idleTimeout {
					log.Printf("Session %s is idle (last active: %v ago), but keeping it active", session.ID, timeSinceLastActive)
					// Don't stop idle sessions automatically - let user activity resume them
					// The time won't be counted during idle period anyway
				}
			}
		}
	}

	log.Println("Finished processing time tracking sessions")
}

