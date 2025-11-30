package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/constants"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/google/uuid"
)

// TimeTrackingService handles time tracking business logic
type TimeTrackingService struct {
	timeTrackingRepo repos.TimeTrackingRepository
	taskRepo         repos.TaskRepository
	idleTimeout      time.Duration // Default: 10 minutes
}

// NewTimeTrackingService creates a new time tracking service
func NewTimeTrackingService(
	timeTrackingRepo repos.TimeTrackingRepository,
	taskRepo repos.TaskRepository,
) *TimeTrackingService {
	return &TimeTrackingService{
		timeTrackingRepo: timeTrackingRepo,
		taskRepo:         taskRepo,
		idleTimeout:      10 * time.Minute, // Default 10 minutes
	}
}

// StartTracking starts a new time tracking session for a task
func (s *TimeTrackingService) StartTracking(ctx context.Context, projectID, taskID, userID string) error {
	// Check if task exists and is in progress
	task, err := s.taskRepo.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	// Check if task status is in_progress
	if task.Status != string(constants.TaskStatusInProgress) {
		return errors.New("can only track time for tasks in progress")
	}

	// Check if there's already an active session for this task and user
	existing, err := s.timeTrackingRepo.GetActiveByTaskAndUser(ctx, projectID, taskID, userID)
	if err != nil {
		return err
	}
	if existing != nil {
		// Session already exists, return success
		return nil
	}

	// Create new session
	now := time.Now()
	session := &models.TimeTrackingSession{
		ID:          uuid.New().String(),
		TaskID:      taskID,
		ProjectID:   projectID,
		UserID:      userID,
		StartTime:   now,
		LastActive:  now,
		TotalMinutes: 0,
		IsActive:    true,
		Created:     now,
	}

	return s.timeTrackingRepo.Add(ctx, session)
}

// StopTracking stops an active time tracking session
func (s *TimeTrackingService) StopTracking(ctx context.Context, projectID, taskID, userID string) error {
	// Get active session
	session, err := s.timeTrackingRepo.GetActiveByTaskAndUser(ctx, projectID, taskID, userID)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("no active tracking session found")
	}

	// Finalize time before stopping
	if err := s.aggregateSessionTime(ctx, session); err != nil {
		return fmt.Errorf("failed to aggregate session time: %w", err)
	}

	// Stop the session
	now := time.Now()
	session.IsActive = false
	session.EndTime = &now
	session.Updated = &now

	return s.timeTrackingRepo.Update(ctx, session)
}

// UpdateActivity updates the last active time for a session
func (s *TimeTrackingService) UpdateActivity(ctx context.Context, projectID, taskID, userID string) error {
	session, err := s.timeTrackingRepo.GetActiveByTaskAndUser(ctx, projectID, taskID, userID)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("no active tracking session found")
	}

	now := time.Now()
	return s.timeTrackingRepo.UpdateActivity(ctx, session.ID, now)
}

// GetActiveSession gets the active session for a task and user
func (s *TimeTrackingService) GetActiveSession(ctx context.Context, projectID, taskID, userID string) (*models.TimeTrackingSession, error) {
	return s.timeTrackingRepo.GetActiveByTaskAndUser(ctx, projectID, taskID, userID)
}

// AggregateTime aggregates time from active sessions and updates TimeSpent entries
func (s *TimeTrackingService) AggregateTime(ctx context.Context, projectID, taskID string) error {
	// Get all active sessions for this task
	sessions, err := s.timeTrackingRepo.GetByTask(ctx, projectID, taskID)
	if err != nil {
		return err
	}

	// Group sessions by user and date, then aggregate
	timeByUserAndDate := make(map[string]map[string]int) // userID -> date -> minutes

	for _, session := range sessions {
		if !session.IsActive {
			continue
		}

		// Calculate time since last activity
		now := time.Now()
		timeSinceLastActive := now.Sub(session.LastActive)

		// If idle (more than idleTimeout), don't count the idle period
		var activeMinutes int
		if timeSinceLastActive > s.idleTimeout {
			// Only count time up to last active
			activeTime := session.LastActive.Sub(session.StartTime)
			activeMinutes = int(activeTime.Minutes()) - session.TotalMinutes
		} else {
			// Count all time since start (minus already aggregated)
			activeTime := now.Sub(session.StartTime)
			activeMinutes = int(activeTime.Minutes()) - session.TotalMinutes
		}

		if activeMinutes > 0 {
			// Get date for the session
			date := session.StartTime.Format("2006-01-02")

			if timeByUserAndDate[session.UserID] == nil {
				timeByUserAndDate[session.UserID] = make(map[string]int)
			}
			timeByUserAndDate[session.UserID][date] += activeMinutes

			// Update session total minutes
			session.TotalMinutes += activeMinutes
			session.LastActive = now
			if err := s.timeTrackingRepo.Update(ctx, &session); err != nil {
				return err
			}
		}
	}

	// Update TimeSpent entries in task
	if len(timeByUserAndDate) > 0 {
		task, err := s.taskRepo.GetByID(ctx, projectID, taskID)
		if err != nil {
			return err
		}
		if task == nil {
			return errors.New("task not found")
		}

		// Merge with existing TimeSpent entries
		for userID, dateMap := range timeByUserAndDate {
			for date, minutes := range dateMap {
				// Find existing entry for this user and date
				found := false
				for i := range task.TimeSpent {
					if task.TimeSpent[i].UserID == userID && task.TimeSpent[i].Date == date {
						task.TimeSpent[i].TimeSpent += minutes
						found = true
						break
					}
				}

				// If not found, create new entry
				if !found {
					task.TimeSpent = append(task.TimeSpent, models.TimeSpent{
						Date:      date,
						TimeSpent: minutes,
						UserID:    userID,
					})
				}
			}
		}

		// Update task
		return s.taskRepo.Update(ctx, task)
	}

	return nil
}

// aggregateSessionTime aggregates time for a single session and updates TimeSpent
func (s *TimeTrackingService) aggregateSessionTime(ctx context.Context, session *models.TimeTrackingSession) error {
	now := time.Now()
	timeSinceLastActive := now.Sub(session.LastActive)

	var activeMinutes int
	if timeSinceLastActive > s.idleTimeout {
		// Only count time up to last active
		activeTime := session.LastActive.Sub(session.StartTime)
		activeMinutes = int(activeTime.Minutes()) - session.TotalMinutes
	} else {
		// Count all time since start
		activeTime := now.Sub(session.StartTime)
		activeMinutes = int(activeTime.Minutes()) - session.TotalMinutes
	}

	if activeMinutes > 0 {
		date := session.StartTime.Format("2006-01-02")

		// Get task and update TimeSpent
		task, err := s.taskRepo.GetByID(ctx, session.ProjectID, session.TaskID)
		if err != nil {
			return err
		}
		if task == nil {
			return errors.New("task not found")
		}

		// Find or create TimeSpent entry
		found := false
		for i := range task.TimeSpent {
			if task.TimeSpent[i].UserID == session.UserID && task.TimeSpent[i].Date == date {
				task.TimeSpent[i].TimeSpent += activeMinutes
				found = true
				break
			}
		}

		if !found {
			task.TimeSpent = append(task.TimeSpent, models.TimeSpent{
				Date:      date,
				TimeSpent: activeMinutes,
				UserID:    session.UserID,
			})
		}

		// Update session
		session.TotalMinutes += activeMinutes
		session.LastActive = now
		if err := s.timeTrackingRepo.Update(ctx, session); err != nil {
			return err
		}

		// Update task
		return s.taskRepo.Update(ctx, task)
	}

	return nil
}

// GetAllActiveSessions gets all active time tracking sessions
func (s *TimeTrackingService) GetAllActiveSessions(ctx context.Context) ([]models.TimeTrackingSession, error) {
	return s.timeTrackingRepo.GetAllActive(ctx)
}

