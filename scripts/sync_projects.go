// sync_projects.go - Syncs projects from oldData/projects.json to the database
//
// This script reads projects from oldData/projects.json and adds them to the database.
// Projects are identified by the pattern: ddmm-title (4 digits followed by a dash).
//
// Usage:
//
//	go run scripts/sync_projects.go
//	go run scripts/sync_projects.go -owner-id "user-id-here"
//	go run scripts/sync_projects.go -dry-run  # Preview without adding to database
//
// Flags:
//
//	-owner-id string   Owner user ID for all projects (if not provided, uses first admin user)
//	-dry-run          Preview mode - shows what would be added without actually adding
//
// Environment Variables Required:
//
//	MONGODB_URI         MongoDB connection string (default: "mongodb://localhost:27017")
//	MONGODB_DATABASE    MongoDB database name (default: "ar13_backend")
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/internal/repos"
	"github.com/ar-13-go-backend/pkg/mongodb"
)

// FileEntry represents an entry from the projects.json file
type FileEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsFolder bool   `json:"isFolder"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
	MimeType string `json:"mimeType,omitempty"`
}

func main() {
	// Parse command-line flags
	ownerID := flag.String("owner-id", "", "Owner user ID for all projects (if not provided, uses first admin user)")
	dryRun := flag.Bool("dry-run", false, "Preview mode - shows what would be added without actually adding")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MongoDB
	_, err = mongodb.InitializeMongoDB(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	fmt.Println("✅ MongoDB initialized successfully")

	ctx := context.Background()

	// Get owner ID
	actualOwnerID, err := getOwnerID(ctx, *ownerID)
	if err != nil {
		log.Fatalf("Failed to get owner ID: %v", err)
	}
	fmt.Printf("✅ Using owner ID: %s\n\n", actualOwnerID)

	// Read and parse JSON file
	jsonPath := filepath.Join("oldData", "projects.json")
	entries, err := readProjectsJSON(jsonPath)
	if err != nil {
		log.Fatalf("Failed to read projects.json: %v", err)
	}

	// Filter and extract projects
	projects := extractProjects(entries)
	fmt.Printf("📋 Found %d valid projects in JSON file\n\n", len(projects))

	if *dryRun {
		fmt.Println("🔍 DRY RUN MODE - No changes will be made\n")
	}

	// Sync projects to database
	synced, skipped, errors := syncProjects(ctx, projects, actualOwnerID, *dryRun)

	// Print summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 SYNC SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("✅ Successfully synced: %d projects\n", synced)
	fmt.Printf("⏭️  Skipped (already exist): %d projects\n", skipped)
	fmt.Printf("❌ Errors: %d projects\n", errors)
	fmt.Println(strings.Repeat("=", 60))

	if errors > 0 {
		os.Exit(1)
	}
}

// readProjectsJSON reads and parses the projects.json file
func readProjectsJSON(path string) ([]FileEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var entries []FileEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return entries, nil
}

// extractProjects filters entries and extracts project information
// Projects must match the pattern: ddmm-title (4 digits followed by a dash)
func extractProjects(entries []FileEntry) []map[string]string {
	projects := []map[string]string{}
	
	// Pattern: 4 digits followed by a dash, then the title
	pattern := regexp.MustCompile(`^(\d{4})-(.+)$`)

	for _, entry := range entries {
		// Only process folders
		if !entry.IsFolder {
			continue
		}

		// Check if name matches project pattern
		matches := pattern.FindStringSubmatch(entry.Name)
		if len(matches) != 3 {
			continue
		}

		code := matches[1]    // ddmm part
		title := strings.TrimSpace(matches[2]) // title part

		// Skip if title is empty
		if title == "" {
			continue
		}

		projects = append(projects, map[string]string{
			"code": code,
			"title": title,
			"name": entry.Name,
		})
	}

	return projects
}

// getOwnerID gets the owner ID for projects
// If ownerID is provided, it uses that. Otherwise, it finds the first admin user.
func getOwnerID(ctx context.Context, ownerID string) (string, error) {
	if ownerID != "" {
		// Validate that the provided owner ID exists
		userRepo := repos.NewUserRepo()
		user, err := userRepo.GetByID(ctx, ownerID)
		if err != nil {
			return "", fmt.Errorf("failed to get user by ID: %w", err)
		}
		if user == nil {
			return "", fmt.Errorf("user with ID %s not found", ownerID)
		}
		return ownerID, nil
	}

	// Find first admin user
	userRepo := repos.NewUserRepo()
	users, err := userRepo.GetAll(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get users: %w", err)
	}

	for _, user := range users {
		if user.Role == models.UserRoleAdmin {
			fmt.Printf("👤 Found admin user: %s (%s)\n", user.Name, user.Email)
			return user.ID, nil
		}
	}

	return "", fmt.Errorf("no admin user found. Please create an admin user first or provide -owner-id flag")
}

// syncProjects syncs projects to the database
func syncProjects(ctx context.Context, projects []map[string]string, ownerID string, dryRun bool) (synced, skipped, errors int) {
	projectRepo := repos.NewProjectRepo()

	for i, projectData := range projects {
		code := projectData["code"]
		title := projectData["title"]
		name := projectData["name"]

		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(projects), name)

		// Check if project already exists (by title or code)
		exists, err := checkProjectExists(ctx, projectRepo, title, code)
		if err != nil {
			fmt.Printf("   ❌ Error checking if project exists: %v\n", err)
			errors++
			continue
		}

		if exists {
			fmt.Printf("   ⏭️  Project already exists, skipping\n")
			skipped++
			continue
		}

		if dryRun {
			fmt.Printf("   🔍 Would add: Code=%s, Title=%s\n", code, title)
			synced++
			continue
		}

		// Create project
		project := &models.Project{
			Title:       title,
			Description: fmt.Sprintf("Project synced from old data: %s", name),
			OwnerID:     ownerID,
			MembersIDs:  []string{},
			Code:        code,
			IsArchived:  false,
		}

		if err := projectRepo.Add(ctx, project); err != nil {
			fmt.Printf("   ❌ Failed to add project: %v\n", err)
			errors++
			continue
		}

		fmt.Printf("   ✅ Added successfully (ID: %s)\n", project.ID)
		synced++
	}

	return synced, skipped, errors
}

// checkProjectExists checks if a project already exists by title or code
func checkProjectExists(ctx context.Context, projectRepo *repos.ProjectRepo, title, code string) (bool, error) {
	// Check by title
	exists, err := projectRepo.Persists(ctx, title)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	// Check by code - we need to query by code
	// Since Persists only checks by title, we'll use GetAll and filter
	projects, err := projectRepo.GetAll(ctx, nil)
	if err != nil {
		return false, err
	}

	for _, p := range projects {
		if p.Code == code {
			return true, nil
		}
	}

	return false, nil
}





