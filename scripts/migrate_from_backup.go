package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/dynamodb"
	"github.com/ar-13-go-backend/pkg/password"
)

// BackupData represents the structure of a backup file
type BackupData struct {
	Collection string           `json:"collection"`
	Documents  []BackupDocument `json:"documents"`
	BackedUpAt time.Time        `json:"backedUpAt"`
}

// BackupDocument represents a document with its data and subcollections
type BackupDocument struct {
	ID             string                 `json:"id"`
	Data           map[string]interface{} `json:"data"`
	SubCollections map[string]BackupData  `json:"subCollections,omitempty"`
}

// Collection to table name mapping
var collectionToTable = map[string]string{
	"users":             "users",
	"projects":          "projects",
	"tasks":             "tasks",
	"notifications":     "notifications",
	"calendar":          "calendar_events",
	"calendar_events":   "calendar_events",
	"vacations":         "leaveRequests",
	"leaveRequests":     "leaveRequests",
	"activityLogs":      "activity_logs",
	"activity_logs":     "activity_logs",
	"infoPortal":        "info-portal",
	"info-portal":       "info-portal",
	"projectDetails":    "project_details",
	"project_details":   "project_details",
	"userAccountLinks":  "userAccountLinks",
	"signupInvitations": "signupInvitations",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/migrate_from_backup.go <backup_directory>")
		fmt.Println("Example: go run scripts/migrate_from_backup.go upload/backups")
		os.Exit(1)
	}

	backupDir := os.Args[1]

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize DynamoDB
	_, err = dynamodb.InitializeDynamoDB(cfg.AWSRegion)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB: %v", err)
	}
	fmt.Println("✅ DynamoDB initialized")

	ctx := context.Background()

	// Migration order matters (users first, then projects, then tasks)
	migrationOrder := []string{
		"users",
		"projects",
		"notifications",
		"calendar_events",
		"leaveRequests",
		"activity_logs",
		"project_details",
		"signupInvitations",
		"userAccountLinks",
		"info-portal",
		"tasks", // Tasks last because they're nested in projects
	}

	// Find backup files
	backupFiles := make(map[string]string)
	err = filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			// Extract collection name from filename
			filename := info.Name()
			// Remove timestamp and .json extension
			parts := strings.Split(filename, "_")
			if len(parts) > 0 {
				collectionName := parts[0]
				backupFiles[collectionName] = path
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to scan backup directory: %v", err)
	}

	fmt.Printf("\n📁 Found %d backup files\n\n", len(backupFiles))

	// Migrate in order
	for _, collectionName := range migrationOrder {
		// Try different filename patterns
		filePath := ""
		for pattern, path := range backupFiles {
			if strings.EqualFold(pattern, collectionName) ||
				strings.EqualFold(pattern, getCollectionName(collectionName)) {
				filePath = path
				break
			}
		}

		if filePath == "" {
			fmt.Printf("⚠️  Skipping %s - backup file not found\n", collectionName)
			continue
		}

		fmt.Printf("📦 Migrating %s from %s...\n", collectionName, filepath.Base(filePath))
		if err := migrateCollection(ctx, collectionName, filePath); err != nil {
			log.Printf("❌ Error migrating %s: %v\n", collectionName, err)
		} else {
			fmt.Printf("✅ Successfully migrated %s\n\n", collectionName)
		}
	}

	// Handle tasks separately (they might be in projects subcollections)
	fmt.Println("📦 Migrating tasks from projects subcollections...")
	if err := migrateTasksFromProjects(ctx, backupDir); err != nil {
		log.Printf("⚠️  Warning: Could not migrate tasks from projects: %v\n", err)
	}

	fmt.Println("\n🎉 Migration complete!")
}

func getCollectionName(tableName string) string {
	// Convert table name to collection name
	mapping := map[string]string{
		"users":             "users",
		"projects":          "projects",
		"tasks":             "tasks",
		"notifications":     "notifications",
		"calendar_events":   "calendar",
		"leaveRequests":     "vacations",
		"activity_logs":     "activityLogs",
		"info-portal":       "infoPortal",
		"project_details":   "projectDetails",
		"userAccountLinks":  "userAccountLinks",
		"signupInvitations": "signupInvitations",
	}
	if name, ok := mapping[tableName]; ok {
		return name
	}
	return tableName
}

func migrateCollection(ctx context.Context, tableName, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var backup BackupData
	if err := json.Unmarshal(data, &backup); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(backup.Documents) == 0 {
		fmt.Printf("   ℹ️  No documents to migrate\n")
		return nil
	}

	switch tableName {
	case "users":
		return migrateUsers(ctx, backup.Documents)
	case "projects":
		return migrateProjects(ctx, backup.Documents)
	case "tasks":
		return migrateTasks(ctx, backup.Documents)
	case "notifications":
		return migrateNotifications(ctx, backup.Documents)
	case "calendar_events":
		return migrateCalendarEvents(ctx, backup.Documents)
	case "leaveRequests":
		return migrateLeaveRequests(ctx, backup.Documents)
	case "activity_logs":
		return migrateActivityLogs(ctx, backup.Documents)
	case "project_details":
		return migrateProjectDetails(ctx, backup.Documents)
	case "signupInvitations":
		return migrateSignupInvitations(ctx, backup.Documents)
	case "userAccountLinks":
		return migrateUserAccountLinks(ctx, backup.Documents)
	case "info-portal":
		return migrateInfoPortal(ctx, backup.Documents)
	default:
		return fmt.Errorf("unknown collection: %s", tableName)
	}
}

func migrateUsers(ctx context.Context, documents []BackupDocument) error {
	userRepo := repos.NewUserRepo()
	successCount := 0

	for _, doc := range documents {
		data := doc.Data
		user := &models.User{
			ID:          getString(data, "id"),
			Name:        getString(data, "name"),
			Email:       getString(data, "email"),
			PhoneNumber: getString(data, "phoneNumber"),
			Role:        models.UserRole(getString(data, "role")),
		}

		// Handle password - check if already hashed
		passwordStr := getString(data, "password")
		if passwordStr != "" {
			// Check if already hashed (bcrypt hashes start with $2a$, $2b$, or $2y$)
			if strings.HasPrefix(passwordStr, "$2a$") || strings.HasPrefix(passwordStr, "$2b$") || strings.HasPrefix(passwordStr, "$2y$") {
				user.Password = passwordStr
			} else {
				// Hash plain text password
				hashed, err := password.HashPassword(passwordStr)
				if err != nil {
					log.Printf("   ⚠️  Failed to hash password for user %s: %v\n", user.Email, err)
					continue
				}
				user.Password = hashed
			}
		}

		// Handle designation
		if designation := getString(data, "designation"); designation != "" {
			user.Designation = &designation
		}

		// Handle timestamps
		if createdStr := getString(data, "created"); createdStr != "" {
			if t, err := parseTime(createdStr); err == nil {
				user.CreatedAt = t
			}
		}
		if createdAtStr := getString(data, "createdAt"); createdAtStr != "" {
			if t, err := parseTime(createdAtStr); err == nil {
				user.CreatedAt = t
			}
		}
		if updatedAtStr := getString(data, "updatedAt"); updatedAtStr != "" {
			if t, err := parseTime(updatedAtStr); err == nil {
				user.UpdatedAt = t
			}
		}

		if err := userRepo.Add(ctx, user); err != nil {
			log.Printf("   ⚠️  Failed to migrate user %s: %v\n", user.Email, err)
			continue
		}
		successCount++
	}

	fmt.Printf("   ✅ Migrated %d/%d users\n", successCount, len(documents))
	return nil
}

func migrateProjects(ctx context.Context, documents []BackupDocument) error {
	projectRepo := repos.NewProjectRepo()
	successCount := 0

	for _, doc := range documents {
		data := doc.Data
		project := &models.Project{
			Model: models.Model{
				ID: getString(data, "id"),
			},
			Title:       getString(data, "title"),
			Description: getString(data, "description"),
			OwnerID:     getString(data, "ownerId"),
		}

		// Handle membersIds
		if members, ok := data["membersIds"].([]interface{}); ok {
			project.MembersIDs = make([]string, 0, len(members))
			for _, m := range members {
				if str, ok := m.(string); ok {
					project.MembersIDs = append(project.MembersIDs, str)
				}
			}
		}

		// Handle deadline
		if deadlineStr := getString(data, "deadLine"); deadlineStr != "" {
			if t, err := parseTime(deadlineStr); err == nil {
				project.Deadline = t
			}
		}

		// Handle logoUrl
		if logoURL := getString(data, "logoUrl"); logoURL != "" {
			project.LogoURL = &logoURL
		}

		// Handle timestamps
		if createdStr := getString(data, "created"); createdStr != "" {
			if t, err := parseTime(createdStr); err == nil {
				project.Created = t
			}
		}

		if err := projectRepo.Add(ctx, project); err != nil {
			log.Printf("   ⚠️  Failed to migrate project %s: %v\n", project.ID, err)
			continue
		}
		successCount++
	}

	fmt.Printf("   ✅ Migrated %d/%d projects\n", successCount, len(documents))
	return nil
}

func migrateTasks(ctx context.Context, documents []BackupDocument) error {
	taskRepo := repos.NewTaskRepo()
	successCount := 0

	for _, doc := range documents {
		data := doc.Data
		task := &models.Task{
			Model: models.Model{
				ID: getString(data, "id"),
			},
			Subject:   getString(data, "subject"),
			Code:      getString(data, "code"),
			Status:    getString(data, "status"),
			Priority:  getString(data, "priority"),
			ProjectID: getString(data, "projectId"),
		}

		// Handle description
		if desc := getString(data, "description"); desc != "" {
			task.Description = &desc
		}

		// Handle deadline (previously duration) - support both for backward compatibility
		if deadlineStr := getString(data, "deadline"); deadlineStr != "" {
			if t, err := parseTime(deadlineStr); err == nil {
				task.Deadline = t
			}
		} else if durationStr := getString(data, "duration"); durationStr != "" {
			// Backward compatibility: migrate old "duration" field to "deadline"
			if t, err := parseTime(durationStr); err == nil {
				task.Deadline = t
			}
		}

		// Handle assignTo - take first element if array exists, otherwise nil
		if assignTo, ok := data["assignTo"].([]interface{}); ok && len(assignTo) > 0 {
			if str, ok := assignTo[0].(string); ok && str != "" {
				task.AssignTo = &str
			}
		}

		// Handle timeSpent
		if timeSpent, ok := data["timeSpent"].([]interface{}); ok {
			task.TimeSpent = make([]models.TimeSpent, 0, len(timeSpent))
			for _, ts := range timeSpent {
				if tsMap, ok := ts.(map[string]interface{}); ok {
					timeSpentItem := models.TimeSpent{
						Date:      getString(tsMap, "date"),
						TimeSpent: getInt(tsMap, "timeSpent"),
						UserID:    getString(tsMap, "userId"),
					}
					if desc := getString(tsMap, "description"); desc != "" {
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
					attachment := models.FileAttachment{
						FileName:     getString(faMap, "fileName"),
						OriginalName: getString(faMap, "originalName"),
						FileSize:     getInt64(faMap, "fileSize"),
						MimeType:     getString(faMap, "mimeType"),
						UploadedBy:   getString(faMap, "uploadedBy"),
						FileURL:      getString(faMap, "fileUrl"),
					}
					if uploadDateStr := getString(faMap, "uploadDate"); uploadDateStr != "" {
						if t, err := parseTime(uploadDateStr); err == nil {
							attachment.UploadDate = t
						}
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
					log := models.ActivityLog{
						ID:          getString(alMap, "id"),
						Type:        models.ActivityType(getString(alMap, "type")),
						UserID:      getString(alMap, "userId"),
						Description: getString(alMap, "description"),
					}
					if timestampStr := getString(alMap, "timestamp"); timestampStr != "" {
						if t, err := parseTime(timestampStr); err == nil {
							log.Timestamp = t
						}
					}
					if userName := getString(alMap, "userName"); userName != "" {
						log.UserName = &userName
					}
					if metadata, ok := alMap["metadata"].(map[string]interface{}); ok {
						log.Metadata = metadata
					}
					task.ActivityLogs = append(task.ActivityLogs, log)
				}
			}
		}

		// Handle timestamps
		if createdStr := getString(data, "created"); createdStr != "" {
			if t, err := parseTime(createdStr); err == nil {
				task.Created = t
			}
		}

		if err := taskRepo.Add(ctx, task); err != nil {
			log.Printf("   ⚠️  Failed to migrate task %s: %v\n", task.ID, err)
			continue
		}
		successCount++
	}

	fmt.Printf("   ✅ Migrated %d/%d tasks\n", successCount, len(documents))
	return nil
}

func migrateNotifications(ctx context.Context, documents []BackupDocument) error {
	notificationRepo := repos.NewNotificationRepo()
	successCount := 0

	for _, doc := range documents {
		data := doc.Data
		notification := &models.Notification{
			Model: models.Model{
				ID: getString(data, "id"),
			},
			Title:             getString(data, "title"),
			Message:           getString(data, "message"),
			Type:              models.NotificationType(getString(data, "type")),
			UserID:            getString(data, "userId"),
			RelatedEntityID:   getString(data, "relatedEntityId"),
			RelatedEntityType: models.RelatedEntityType(getString(data, "relatedEntityType")),
			IsRead:            getBool(data, "isRead"),
		}

		// Handle timestamps
		if createdAtStr := getString(data, "createdAt"); createdAtStr != "" {
			if t, err := parseTime(createdAtStr); err == nil {
				notification.CreatedAt = t
			}
		}
		if createdStr := getString(data, "created"); createdStr != "" {
			if t, err := parseTime(createdStr); err == nil {
				notification.Created = t
			}
		}

		if err := notificationRepo.Add(ctx, notification); err != nil {
			log.Printf("   ⚠️  Failed to migrate notification %s: %v\n", notification.ID, err)
			continue
		}
		successCount++
	}

	fmt.Printf("   ✅ Migrated %d/%d notifications\n", successCount, len(documents))
	return nil
}

func migrateCalendarEvents(ctx context.Context, documents []BackupDocument) error {
	calendarRepo := repos.NewCalendarEventRepo()
	successCount := 0

	for _, doc := range documents {
		data := doc.Data
		event := &models.CalendarEvent{
			Model: models.Model{
				ID: getString(data, "id"),
			},
			Title:       getString(data, "title"),
			Category:    getString(data, "category"),
			Priority:    getString(data, "priority"),
			CreatedBy:   getString(data, "createdBy"),
			IsRepeating: getBool(data, "isRepeating"),
		}

		// Handle start/end dates
		if startStr := getString(data, "start"); startStr != "" {
			if t, err := parseTime(startStr); err == nil {
				event.Start = t
			}
		}
		if endStr := getString(data, "end"); endStr != "" {
			if t, err := parseTime(endStr); err == nil {
				event.End = t
			}
		}

		// Handle optional fields
		if timeStr := getString(data, "time"); timeStr != "" {
			event.Time = &timeStr
		}
		if desc := getString(data, "description"); desc != "" {
			event.Description = &desc
		}
		if freq := getString(data, "repeatFrequency"); freq != "" {
			f := models.RepeatFrequency(freq)
			event.RepeatFrequency = &f
		}
		if repeatDays, ok := data["repeatDays"].([]interface{}); ok {
			days := make([]string, 0, len(repeatDays))
			for _, d := range repeatDays {
				if str, ok := d.(string); ok {
					days = append(days, str)
				}
			}
			event.RepeatDays = days
		}
		if invites, ok := data["invites"].([]interface{}); ok {
			inviteEmails := make([]string, 0, len(invites))
			for _, invite := range invites {
				if str, ok := invite.(string); ok {
					inviteEmails = append(inviteEmails, str)
				}
			}
			if len(inviteEmails) > 0 {
				event.Invites = inviteEmails
			}
		}

		// Handle timestamps
		if createdStr := getString(data, "created"); createdStr != "" {
			if t, err := parseTime(createdStr); err == nil {
				event.Created = t
			}
		}

		if err := calendarRepo.Add(ctx, event); err != nil {
			log.Printf("   ⚠️  Failed to migrate calendar event %s: %v\n", event.ID, err)
			continue
		}
		successCount++
	}

	fmt.Printf("   ✅ Migrated %d/%d calendar events\n", successCount, len(documents))
	return nil
}

func migrateLeaveRequests(ctx context.Context, documents []BackupDocument) error {
	vacationRepo := repos.NewVacationRepo()
	successCount := 0

	for _, doc := range documents {
		data := doc.Data
		request := &models.LeaveRequest{
			Model: models.Model{
				ID: getString(data, "id"),
			},
			UserID:       getString(data, "userId"),
			RequestType:  models.LeaveRequestType(getString(data, "requestType")),
			Duration:     getFloat64(data, "duration"),
			DurationType: models.DurationType(getString(data, "durationType")),
			Status:       models.LeaveRequestStatus(getString(data, "status")),
		}

		// Handle dates
		if startDateStr := getString(data, "startDate"); startDateStr != "" {
			if t, err := parseTime(startDateStr); err == nil {
				request.StartDate = t
			}
		}
		if endDateStr := getString(data, "endDate"); endDateStr != "" {
			if t, err := parseTime(endDateStr); err == nil {
				request.EndDate = &t
			}
		}
		if requestedAtStr := getString(data, "requestedAt"); requestedAtStr != "" {
			if t, err := parseTime(requestedAtStr); err == nil {
				request.RequestedAt = t
			}
		}
		if reviewedAtStr := getString(data, "reviewedAt"); reviewedAtStr != "" {
			if t, err := parseTime(reviewedAtStr); err == nil {
				request.ReviewedAt = &t
			}
		}

		// Handle optional fields
		if comments := getString(data, "comments"); comments != "" {
			request.Comments = &comments
		}
		if reviewedBy := getString(data, "reviewedBy"); reviewedBy != "" {
			request.ReviewedBy = &reviewedBy
		}
		if reviewComments := getString(data, "reviewComments"); reviewComments != "" {
			request.ReviewComments = &reviewComments
		}
		if workingHours, ok := data["workingHours"].(map[string]interface{}); ok {
			wh := models.WorkingHours{
				From: getString(workingHours, "from"),
				To:   getString(workingHours, "to"),
			}
			request.WorkingHours = &wh
		}

		// Handle timestamps
		if createdStr := getString(data, "created"); createdStr != "" {
			if t, err := parseTime(createdStr); err == nil {
				request.Created = t
			}
		}

		if err := vacationRepo.Add(ctx, request); err != nil {
			log.Printf("   ⚠️  Failed to migrate leave request %s: %v\n", request.ID, err)
			continue
		}
		successCount++
	}

	fmt.Printf("   ✅ Migrated %d/%d leave requests\n", successCount, len(documents))
	return nil
}

func migrateActivityLogs(ctx context.Context, documents []BackupDocument) error {
	activityLogRepo := repos.NewActivityLogRepo()
	successCount := 0

	for _, doc := range documents {
		data := doc.Data
		log := &models.ActivityLogBase{
			Model: models.Model{
				ID: getString(data, "id"),
			},
			EntityType: models.ActivityLogEntityType(getString(data, "entityType")),
			EntityID:   getString(data, "entityId"),
			Action:     models.ActivityLogAction(getString(data, "action")),
			CreatedBy:  getString(data, "createdBy"),
		}

		// Handle optional fields
		if desc := getString(data, "description"); desc != "" {
			log.Description = &desc
		}
		if fields, ok := data["fields"].(map[string]interface{}); ok {
			log.Fields = fields
		}
		if metadata, ok := data["metadata"].(map[string]interface{}); ok {
			log.Metadata = metadata
		}

		// Handle timestamps
		if createdAtStr := getString(data, "createdAt"); createdAtStr != "" {
			if t, err := parseTime(createdAtStr); err == nil {
				log.CreatedAt = t
			}
		}
		if createdStr := getString(data, "created"); createdStr != "" {
			if t, err := parseTime(createdStr); err == nil {
				log.Created = t
			}
		}

		if err := activityLogRepo.Add(ctx, log); err != nil {
			fmt.Printf("   ⚠️  Failed to migrate activity log %s: %v\n", log.ID, err)
			continue
		}
		successCount++
	}

	fmt.Printf("   ✅ Migrated %d/%d activity logs\n", successCount, len(documents))
	return nil
}

func migrateProjectDetails(ctx context.Context, documents []BackupDocument) error {
	projectDetailsRepo := repos.NewProjectDetailsRepo()
	successCount := 0

	for _, doc := range documents {
		data := doc.Data
		projectID := getString(data, "projectId")
		if projectID == "" {
			continue
		}

		details := &models.ProjectDetails{
			Model: models.Model{
				ID: getString(data, "id"),
			},
			ProjectID: projectID,
		}

		// Handle data field
		if dataField, ok := data["data"].(map[string]interface{}); ok {
			details.Data = dataField
		}

		// Handle timestamps
		if createdStr := getString(data, "created"); createdStr != "" {
			if t, err := parseTime(createdStr); err == nil {
				details.Created = t
			}
		}

		if err := projectDetailsRepo.Add(ctx, projectID, details); err != nil {
			log.Printf("   ⚠️  Failed to migrate project details %s: %v\n", details.ID, err)
			continue
		}
		successCount++
	}

	fmt.Printf("   ✅ Migrated %d/%d project details\n", successCount, len(documents))
	return nil
}

func migrateSignupInvitations(ctx context.Context, documents []BackupDocument) error {
	signupRepo := repos.NewSignupInvitationRepo()
	successCount := 0

	for _, doc := range documents {
		data := doc.Data
		invitation := &models.SignupInvitation{
			ID:        getString(data, "id"),
			Email:     getString(data, "email"),
			Token:     getString(data, "token"),
			HasSignup: getBool(data, "hasSignup"),
		}

		// Handle linkExpiry
		if linkExpiryStr := getString(data, "linkExpiry"); linkExpiryStr != "" {
			if t, err := parseTime(linkExpiryStr); err == nil {
				invitation.LinkExpiry = t
			}
		}

		// Handle timestamps
		if createdStr := getString(data, "created"); createdStr != "" {
			if t, err := parseTime(createdStr); err == nil {
				invitation.Created = t
			}
		}

		if err := signupRepo.Add(ctx, invitation); err != nil {
			log.Printf("   ⚠️  Failed to migrate signup invitation %s: %v\n", invitation.ID, err)
			continue
		}
		successCount++
	}

	fmt.Printf("   ✅ Migrated %d/%d signup invitations\n", successCount, len(documents))
	return nil
}

func migrateUserAccountLinks(ctx context.Context, documents []BackupDocument) error {
	accountLinkRepo := repos.NewUserAccountLinkRepo()
	successCount := 0

	for _, doc := range documents {
		data := doc.Data
		link := &models.UserAccountLink{
			Model: models.Model{
				ID: getString(data, "id"),
			},
			UserID:         getString(data, "userId"),
			Provider:       models.AccountProvider(getString(data, "provider")),
			ProviderUserID: getString(data, "providerUserId"),
			ProviderEmail:  getString(data, "providerEmail"),
			IsActive:       getBool(data, "isActive"),
		}

		// Handle optional fields
		if displayName := getString(data, "providerDisplayName"); displayName != "" {
			link.ProviderDisplayName = &displayName
		}
		if accessToken := getString(data, "accessToken"); accessToken != "" {
			link.AccessToken = &accessToken
		}
		if refreshToken := getString(data, "refreshToken"); refreshToken != "" {
			link.RefreshToken = &refreshToken
		}
		if expiresAtStr := getString(data, "expiresAt"); expiresAtStr != "" {
			if t, err := parseTime(expiresAtStr); err == nil {
				link.ExpiresAt = &t
			}
		}

		// Handle timestamps
		if linkedAtStr := getString(data, "linkedAt"); linkedAtStr != "" {
			if t, err := parseTime(linkedAtStr); err == nil {
				link.LinkedAt = t
			}
		}
		if createdStr := getString(data, "created"); createdStr != "" {
			if t, err := parseTime(createdStr); err == nil {
				link.Created = t
			}
		}

		if err := accountLinkRepo.Add(ctx, link); err != nil {
			log.Printf("   ⚠️  Failed to migrate account link %s: %v\n", link.ID, err)
			continue
		}
		successCount++
	}

	fmt.Printf("   ✅ Migrated %d/%d user account links\n", successCount, len(documents))
	return nil
}

func migrateInfoPortal(ctx context.Context, documents []BackupDocument) error {
	infoPortalRepo := repos.NewInfoPortalRepo()
	successCount := 0

	for _, doc := range documents {
		data := doc.Data
		itemType := getString(data, "type")

		switch itemType {
		case "folder":
			folder := &models.Folder{
				Model: models.Model{
					ID: getString(data, "id"),
				},
				Name:  getString(data, "name"),
				Color: getString(data, "color"),
			}
			if createdStr := getString(data, "created"); createdStr != "" {
				if t, err := parseTime(createdStr); err == nil {
					folder.Created = t
				}
			}
			if err := infoPortalRepo.CreateFolder(ctx, folder); err != nil {
				log.Printf("   ⚠️  Failed to migrate folder %s: %v\n", folder.ID, err)
				continue
			}
			successCount++
		case "page":
			page := &models.Page{
				Model: models.Model{
					ID: getString(data, "id"),
				},
				Title:    getString(data, "title"),
				IsActive: getBool(data, "isActive"),
				FolderID: getString(data, "folderId"),
			}
			if createdStr := getString(data, "created"); createdStr != "" {
				if t, err := parseTime(createdStr); err == nil {
					page.Created = t
				}
			}
			if err := infoPortalRepo.CreatePage(ctx, page.FolderID, page); err != nil {
				log.Printf("   ⚠️  Failed to migrate page %s: %v\n", page.ID, err)
				continue
			}
			successCount++
		case "attachment":
			attachment := &models.Attachment{
				Model: models.Model{
					ID: getString(data, "id"),
				},
				Name:     getString(data, "name"),
				ImageURL: getString(data, "imageUrl"),
				FileURL:  getString(data, "fileUrl"),
				FileType: getString(data, "fileType"),
				FileSize: getInt64(data, "fileSize"),
				PageID:   getString(data, "pageId"),
			}
			if createdStr := getString(data, "created"); createdStr != "" {
				if t, err := parseTime(createdStr); err == nil {
					attachment.Created = t
				}
			}
			if err := infoPortalRepo.CreateAttachment(ctx, attachment.PageID, attachment); err != nil {
				log.Printf("   ⚠️  Failed to migrate attachment %s: %v\n", attachment.ID, err)
				continue
			}
			successCount++
		}
	}

	fmt.Printf("   ✅ Migrated %d info portal items\n", successCount)
	return nil
}

func migrateTasksFromProjects(ctx context.Context, backupDir string) error {
	// Find projects backup file
	projectsFile := ""
	err := filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(info.Name(), "projects") && strings.HasSuffix(path, ".json") {
			projectsFile = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil || projectsFile == "" {
		return fmt.Errorf("projects backup file not found")
	}

	data, err := os.ReadFile(projectsFile)
	if err != nil {
		return err
	}

	var backup BackupData
	if err := json.Unmarshal(data, &backup); err != nil {
		return err
	}

	taskCount := 0

	for _, projectDoc := range backup.Documents {
		if tasks, ok := projectDoc.SubCollections["tasks"]; ok {
			if err := migrateTasks(ctx, tasks.Documents); err == nil {
				taskCount += len(tasks.Documents)
			}
		}
	}

	if taskCount > 0 {
		fmt.Printf("   ✅ Migrated %d tasks from projects subcollections\n", taskCount)
	}

	return nil
}

// Helper functions
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

func getInt64(data map[string]interface{}, key string) int64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}

func getFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0
}

func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func parseTime(timeStr string) (time.Time, error) {
	// Try different time formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}
