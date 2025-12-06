package repos

import (
	"context"
	"errors"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TimeTrackingRepo handles time tracking session data operations with MongoDB
type TimeTrackingRepo struct {
	*MongoBaseRepo
}

// NewTimeTrackingRepo creates a new MongoDB time tracking repository
func NewTimeTrackingRepo() *TimeTrackingRepo {
	client := mongodb.GetClient()
	if client == nil {
		panic("MongoDB client is not initialized. Please ensure MongoDB is connected before creating repositories.")
	}
	return &TimeTrackingRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, mongodb.GetDatabaseName(), "time_tracking_sessions"),
	}
}

// Add creates a new time tracking session
func (r *TimeTrackingRepo) Add(ctx context.Context, session *models.TimeTrackingSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.Created.IsZero() {
		session.Created = time.Now()
	}
	now := time.Now()
	session.Updated = &now

	return r.InsertOne(ctx, session)
}

// GetByID gets a time tracking session by ID
func (r *TimeTrackingRepo) GetByID(ctx context.Context, sessionID string) (*models.TimeTrackingSession, error) {
	filter := bson.M{"id": sessionID}
	result := r.FindOne(ctx, filter)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var session models.TimeTrackingSession
	if err := result.Decode(&session); err != nil {
		return nil, err
	}

	return &session, nil
}

// GetActiveByTaskAndUser gets the active session for a task and user (including paused sessions)
func (r *TimeTrackingRepo) GetActiveByTaskAndUser(ctx context.Context, projectID, taskID, userID string) (*models.TimeTrackingSession, error) {
	filter := bson.M{
		"projectId": projectID,
		"taskId":    taskID,
		"userId":    userID,
		"isActive":  true,
	}
	result := r.FindOne(ctx, filter)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var session models.TimeTrackingSession
	if err := result.Decode(&session); err != nil {
		return nil, err
	}

	return &session, nil
}

// GetAllActiveByUser gets all active sessions for a user (including paused sessions)
func (r *TimeTrackingRepo) GetAllActiveByUser(ctx context.Context, userID string) ([]models.TimeTrackingSession, error) {
	filter := bson.M{
		"userId":   userID,
		"isActive": true,
	}
	results, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	var sessions []models.TimeTrackingSession
	for _, result := range results {
		bsonBytes, err := bson.Marshal(result)
		if err != nil {
			continue
		}
		var session models.TimeTrackingSession
		if err := bson.Unmarshal(bsonBytes, &session); err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetAllActive gets all active time tracking sessions
func (r *TimeTrackingRepo) GetAllActive(ctx context.Context) ([]models.TimeTrackingSession, error) {
	filter := bson.M{"isActive": true}
	results, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	var sessions []models.TimeTrackingSession
	for _, result := range results {
		bsonBytes, err := bson.Marshal(result)
		if err != nil {
			continue
		}
		var session models.TimeTrackingSession
		if err := bson.Unmarshal(bsonBytes, &session); err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetByTask gets all sessions for a task (active and inactive)
func (r *TimeTrackingRepo) GetByTask(ctx context.Context, projectID, taskID string) ([]models.TimeTrackingSession, error) {
	filter := bson.M{
		"projectId": projectID,
		"taskId":    taskID,
	}
	results, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	var sessions []models.TimeTrackingSession
	for _, result := range results {
		bsonBytes, err := bson.Marshal(result)
		if err != nil {
			continue
		}
		var session models.TimeTrackingSession
		if err := bson.Unmarshal(bsonBytes, &session); err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Update updates a time tracking session
func (r *TimeTrackingRepo) Update(ctx context.Context, session *models.TimeTrackingSession) error {
	if session.ID == "" {
		return errors.New("session ID is required")
	}
	now := time.Now()
	session.Updated = &now

	// Convert session to bson.M for update
	update := bson.M{
		"taskId":       session.TaskID,
		"projectId":    session.ProjectID,
		"userId":       session.UserID,
		"startTime":    session.StartTime,
		"lastActive":   session.LastActive,
		"totalMinutes": session.TotalMinutes,
		"isActive":     session.IsActive,
		"isPaused":     session.IsPaused,
		"updated":      session.Updated,
	}
	if session.EndTime != nil {
		update["endTime"] = session.EndTime
	}
	if session.PausedAt != nil {
		update["pausedAt"] = session.PausedAt
	}
	return r.UpdateOne(ctx, session.ID, update)
}

// StopSession stops an active session
func (r *TimeTrackingRepo) StopSession(ctx context.Context, sessionID string) error {
	now := time.Now()
	updates := bson.M{
		"isActive": false,
		"endTime":  now,
		"updated":  now,
	}
	return r.UpdateOne(ctx, sessionID, updates)
}

// UpdateActivity updates the last active time for a session
func (r *TimeTrackingRepo) UpdateActivity(ctx context.Context, sessionID string, lastActive time.Time) error {
	now := time.Now()
	updates := bson.M{
		"lastActive": lastActive,
		"updated":    now,
	}
	return r.UpdateOne(ctx, sessionID, updates)
}

// UpdateTotalMinutes updates the total minutes for a session
func (r *TimeTrackingRepo) UpdateTotalMinutes(ctx context.Context, sessionID string, totalMinutes int) error {
	now := time.Now()
	updates := bson.M{
		"totalMinutes": totalMinutes,
		"updated":      now,
	}
	return r.UpdateOne(ctx, sessionID, updates)
}

