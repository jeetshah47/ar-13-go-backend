package repos

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/firebase"
	"google.golang.org/api/iterator"
)

// TaskRepo handles task data operations
type TaskRepo struct {
	*BaseRepo
}

// NewTaskRepo creates a new task repository
func NewTaskRepo() *TaskRepo {
	return &TaskRepo{
		BaseRepo: NewBaseRepo("tasks"), // Will use subcollection
	}
}

// getCollection gets the tasks collection for a project
func (r *TaskRepo) getCollection(projectID string) *firestore.CollectionRef {
	return firebase.GetCollection("projects").Doc(projectID).Collection("tasks")
}

// dataToTask converts a data map to a Task struct
func (r *TaskRepo) dataToTask(data map[string]interface{}, id string) (*models.Task, error) {
	task := &models.Task{}
	task.ID = id

	// Convert time fields from strings/timestamps to time.Time
	timeFields := []string{"duration", "created", "updated"}
	if err := ConvertTimeFieldsInMap(data, timeFields); err != nil {
		return nil, err
	}

	// Populate basic fields
	if subject, ok := data["subject"].(string); ok {
		task.Subject = subject
	}
	if code, ok := data["code"].(string); ok {
		task.Code = code
	}
	if status, ok := data["status"].(string); ok {
		task.Status = status
	}
	if priority, ok := data["priority"].(string); ok {
		task.Priority = priority
	}
	if projectID, ok := data["projectId"].(string); ok {
		task.ProjectID = projectID
	}
	if description, ok := data["description"].(string); ok {
		task.Description = &description
	}

	// Handle duration
	if durationVal, ok := data["duration"]; ok && durationVal != nil {
		if duration, err := ConvertToTime(durationVal); err == nil {
			task.Duration = duration
		}
	}

	// Handle created
	if createdVal, ok := data["created"]; ok && createdVal != nil {
		if created, err := ConvertToTime(createdVal); err == nil {
			task.Created = created
		}
	}

	// Handle updated
	if updatedVal, ok := data["updated"]; ok && updatedVal != nil {
		if updated, err := ConvertToTime(updatedVal); err == nil {
			task.Updated = &updated
		}
	}

	// Handle assignTo
	if assignTo, ok := data["assignTo"].([]interface{}); ok {
		task.AssignTo = make([]string, 0, len(assignTo))
		for _, id := range assignTo {
			if strID, ok := id.(string); ok {
				task.AssignTo = append(task.AssignTo, strID)
			}
		}
	}

	// Handle timeSpent
	if timeSpent, ok := data["timeSpent"].([]interface{}); ok {
		task.TimeSpent = make([]models.TimeSpent, 0, len(timeSpent))
		for _, ts := range timeSpent {
			if tsMap, ok := ts.(map[string]interface{}); ok {
				timeSpentItem := models.TimeSpent{}
				if date, ok := tsMap["date"].(string); ok {
					timeSpentItem.Date = date
				}
				if timeSpentVal, ok := tsMap["timeSpent"].(int64); ok {
					timeSpentItem.TimeSpent = int(timeSpentVal)
				} else if timeSpentVal, ok := tsMap["timeSpent"].(int); ok {
					timeSpentItem.TimeSpent = timeSpentVal
				}
				if userID, ok := tsMap["userId"].(string); ok {
					timeSpentItem.UserID = userID
				}
				if desc, ok := tsMap["description"].(string); ok {
					timeSpentItem.Description = &desc
				}
				task.TimeSpent = append(task.TimeSpent, timeSpentItem)
			}
		}
	}

	// Handle fileAttachments
	if fileAttachments, ok := data["fileAttachments"].([]interface{}); ok {
		task.FileAttachments = make([]models.FileAttachment, 0, len(fileAttachments))
		for _, fa := range fileAttachments {
			if faMap, ok := fa.(map[string]interface{}); ok {
				// Convert uploadDate
				if err := ConvertTimeFieldsInMap(faMap, []string{"uploadDate"}); err != nil {
					return nil, err
				}

				attachment := models.FileAttachment{}
				if fileName, ok := faMap["fileName"].(string); ok {
					attachment.FileName = fileName
				}
				if originalName, ok := faMap["originalName"].(string); ok {
					attachment.OriginalName = originalName
				}
				if fileSize, ok := faMap["fileSize"].(int64); ok {
					attachment.FileSize = fileSize
				}
				if mimeType, ok := faMap["mimeType"].(string); ok {
					attachment.MimeType = mimeType
				}
				if uploadDateVal, ok := faMap["uploadDate"]; ok && uploadDateVal != nil {
					if uploadDate, err := ConvertToTime(uploadDateVal); err == nil {
						attachment.UploadDate = uploadDate
					}
				}
				if uploadedBy, ok := faMap["uploadedBy"].(string); ok {
					attachment.UploadedBy = uploadedBy
				}
				if fileURL, ok := faMap["fileUrl"].(string); ok {
					attachment.FileURL = fileURL
				}
				task.FileAttachments = append(task.FileAttachments, attachment)
			}
		}
	}

	// Handle activityLogs
	if activityLogs, ok := data["activityLogs"].([]interface{}); ok {
		task.ActivityLogs = make([]models.ActivityLog, 0, len(activityLogs))
		for _, al := range activityLogs {
			if alMap, ok := al.(map[string]interface{}); ok {
				// Convert timestamp
				if err := ConvertTimeFieldsInMap(alMap, []string{"timestamp"}); err != nil {
					return nil, err
				}

				log := models.ActivityLog{}
				if logID, ok := alMap["id"].(string); ok {
					log.ID = logID
				}
				if logType, ok := alMap["type"].(string); ok {
					log.Type = models.ActivityType(logType)
				}
				if timestampVal, ok := alMap["timestamp"]; ok && timestampVal != nil {
					if timestamp, err := ConvertToTime(timestampVal); err == nil {
						log.Timestamp = timestamp
					}
				}
				if userID, ok := alMap["userId"].(string); ok {
					log.UserID = userID
				}
				if userName, ok := alMap["userName"].(string); ok {
					log.UserName = &userName
				}
				if description, ok := alMap["description"].(string); ok {
					log.Description = description
				}
				if metadata, ok := alMap["metadata"].(map[string]interface{}); ok {
					log.Metadata = metadata
				}
				task.ActivityLogs = append(task.ActivityLogs, log)
			}
		}
	}

	return task, nil
}

// GetByID gets a task by ID
func (r *TaskRepo) GetByID(ctx context.Context, projectID, taskID string) (*models.Task, error) {
	doc, err := r.getCollection(projectID).Doc(taskID).Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, nil
	}

	data := doc.Data()
	task, err := r.dataToTask(data, doc.Ref.ID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// GetAll gets all tasks for a project
func (r *TaskRepo) GetAll(ctx context.Context, projectID string) ([]models.Task, error) {
	iter := r.getCollection(projectID).Documents(ctx)
	var tasks []models.Task

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		task, err := r.dataToTask(data, doc.Ref.ID)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

// Add creates a new task
func (r *TaskRepo) Add(ctx context.Context, task *models.Task) error {
	newDocRef := r.getCollection(task.ProjectID).NewDoc()
	task.ID = newDocRef.ID
	task.Created = time.Now()

	if task.TimeSpent == nil {
		task.TimeSpent = []models.TimeSpent{}
	}
	if task.FileAttachments == nil {
		task.FileAttachments = []models.FileAttachment{}
	}
	if task.ActivityLogs == nil {
		task.ActivityLogs = []models.ActivityLog{}
	}

	data := map[string]interface{}{
		"id":              task.ID,
		"subject":         task.Subject,
		"code":            task.Code,
		"status":          task.Status,
		"duration":        task.Duration,
		"priority":        task.Priority,
		"assignTo":        task.AssignTo,
		"projectId":       task.ProjectID,
		"timeSpent":       task.TimeSpent,
		"fileAttachments": task.FileAttachments,
		"activityLogs":    task.ActivityLogs,
		"created":         task.Created,
	}
	if task.Description != nil {
		data["description"] = *task.Description
	}

	_, err := newDocRef.Set(ctx, data)
	return err
}

// Update updates a task
func (r *TaskRepo) Update(ctx context.Context, task *models.Task) error {
	now := time.Now()
	task.Updated = &now

	data := map[string]interface{}{
		"id":              task.ID,
		"subject":         task.Subject,
		"code":            task.Code,
		"status":          task.Status,
		"duration":        task.Duration,
		"priority":        task.Priority,
		"assignTo":        task.AssignTo,
		"projectId":       task.ProjectID,
		"timeSpent":       task.TimeSpent,
		"fileAttachments": task.FileAttachments,
		"activityLogs":    task.ActivityLogs,
		"updated":         now,
	}
	if task.Description != nil {
		data["description"] = *task.Description
	}

	_, err := r.getCollection(task.ProjectID).Doc(task.ID).Set(ctx, data)
	return err
}

// Delete deletes a task
func (r *TaskRepo) Delete(ctx context.Context, projectID, taskID string) error {
	_, err := r.getCollection(projectID).Doc(taskID).Delete(ctx)
	return err
}

// UpdateDuration updates task duration
func (r *TaskRepo) UpdateDuration(ctx context.Context, projectID, taskID string, duration time.Time) error {
	_, err := r.getCollection(projectID).Doc(taskID).Update(ctx, []firestore.Update{
		{Path: "duration", Value: duration},
		{Path: "updated", Value: time.Now()},
	})
	return err
}

// UpdateDescription updates task description
func (r *TaskRepo) UpdateDescription(ctx context.Context, projectID, taskID string, description string) error {
	_, err := r.getCollection(projectID).Doc(taskID).Update(ctx, []firestore.Update{
		{Path: "description", Value: description},
		{Path: "updated", Value: time.Now()},
	})
	return err
}

// UpdateStatus updates task status
func (r *TaskRepo) UpdateStatus(ctx context.Context, projectID, taskID, status string) error {
	_, err := r.getCollection(projectID).Doc(taskID).Update(ctx, []firestore.Update{
		{Path: "status", Value: status},
		{Path: "updated", Value: time.Now()},
	})
	return err
}

// AddTimeSpent adds a time spent entry
func (r *TaskRepo) AddTimeSpent(ctx context.Context, projectID, taskID string, timeSpent models.TimeSpent) error {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	task.TimeSpent = append(task.TimeSpent, timeSpent)
	_, err = r.getCollection(projectID).Doc(taskID).Update(ctx, []firestore.Update{
		{Path: "timeSpent", Value: task.TimeSpent},
		{Path: "updated", Value: time.Now()},
	})
	return err
}

// UpdateTimeSpent updates a time spent entry
func (r *TaskRepo) UpdateTimeSpent(ctx context.Context, projectID, taskID string, index int, timeSpent models.TimeSpent) error {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if index < 0 || index >= len(task.TimeSpent) {
		return errors.New("invalid time spent index")
	}

	task.TimeSpent[index] = timeSpent
	_, err = r.getCollection(projectID).Doc(taskID).Update(ctx, []firestore.Update{
		{Path: "timeSpent", Value: task.TimeSpent},
		{Path: "updated", Value: time.Now()},
	})
	return err
}

// RemoveTimeSpent removes a time spent entry
func (r *TaskRepo) RemoveTimeSpent(ctx context.Context, projectID, taskID string, index int) error {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if index < 0 || index >= len(task.TimeSpent) {
		return errors.New("invalid time spent index")
	}

	task.TimeSpent = append(task.TimeSpent[:index], task.TimeSpent[index+1:]...)
	_, err = r.getCollection(projectID).Doc(taskID).Update(ctx, []firestore.Update{
		{Path: "timeSpent", Value: task.TimeSpent},
		{Path: "updated", Value: time.Now()},
	})
	return err
}

// AddFileAttachment adds a file attachment
func (r *TaskRepo) AddFileAttachment(ctx context.Context, projectID, taskID string, attachment models.FileAttachment) error {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	task.FileAttachments = append(task.FileAttachments, attachment)
	_, err = r.getCollection(projectID).Doc(taskID).Update(ctx, []firestore.Update{
		{Path: "fileAttachments", Value: task.FileAttachments},
		{Path: "updated", Value: time.Now()},
	})
	return err
}

// RemoveFileAttachment removes a file attachment
func (r *TaskRepo) RemoveFileAttachment(ctx context.Context, projectID, taskID string, index int) error {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if index < 0 || index >= len(task.FileAttachments) {
		return errors.New("invalid file attachment index")
	}

	task.FileAttachments = append(task.FileAttachments[:index], task.FileAttachments[index+1:]...)
	_, err = r.getCollection(projectID).Doc(taskID).Update(ctx, []firestore.Update{
		{Path: "fileAttachments", Value: task.FileAttachments},
		{Path: "updated", Value: time.Now()},
	})
	return err
}

// GetTimeSpent gets time spent entries
func (r *TaskRepo) GetTimeSpent(ctx context.Context, projectID, taskID string) ([]models.TimeSpent, error) {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	return task.TimeSpent, nil
}

// GetFileAttachments gets file attachments
func (r *TaskRepo) GetFileAttachments(ctx context.Context, projectID, taskID string) ([]models.FileAttachment, error) {
	task, err := r.GetByID(ctx, projectID, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	return task.FileAttachments, nil
}
