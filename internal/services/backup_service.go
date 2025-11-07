package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ar-13-go-backend/pkg/firebase"
	"google.golang.org/api/iterator"
)

// BackupService handles Firestore backup operations
type BackupService struct {
	client *firestore.Client
}

// NewBackupService creates a new backup service
func NewBackupService() *BackupService {
	return &BackupService{
		client: firebase.GetFirestoreClient(),
	}
}

// BackupData represents the structure of a backup
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

// BackupAllCollections backs up all Firestore collections to JSON files
func (s *BackupService) BackupAllCollections(ctx context.Context, backupDir string) (map[string]string, error) {
	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Get all collections
	collections, err := s.getAllCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collections: %w", err)
	}

	backupFiles := make(map[string]string)
	timestamp := time.Now().Format("20060102_150405")

	// Backup each collection
	for _, collectionName := range collections {
		backupData, err := s.backupCollection(ctx, collectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to backup collection %s: %w", collectionName, err)
		}

		// Save to JSON file
		fileName := fmt.Sprintf("%s_%s.json", collectionName, timestamp)
		filePath := filepath.Join(backupDir, fileName)

		fileData, err := json.MarshalIndent(backupData, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal backup data for %s: %w", collectionName, err)
		}

		if err := os.WriteFile(filePath, fileData, 0644); err != nil {
			return nil, fmt.Errorf("failed to write backup file for %s: %w", collectionName, err)
		}

		backupFiles[collectionName] = filePath
	}

	return backupFiles, nil
}

// getAllCollections gets all top-level collections in Firestore
func (s *BackupService) getAllCollections(ctx context.Context) ([]string, error) {
	// Firestore doesn't have a direct API to list all collections
	// We'll use a known list of collections based on the codebase
	// All known collections will be backed up, even if empty

	// Known collections from the codebase
	knownCollections := []string{
		"users",
		"projects",
		"calendar",
		"notifications",
		"vacation",
		"activityLogs",
		"infoPortal",
		"signupInvitations",
		"userAccountLinks",
	}

	// Return all known collections - Firestore collections are created on first write
	// but we'll backup them all to ensure completeness
	return knownCollections, nil
}

// backupCollection backs up a collection and all its documents
func (s *BackupService) backupCollection(ctx context.Context, collectionName string) (*BackupData, error) {
	collection := s.client.Collection(collectionName)
	iter := collection.Documents(ctx)
	defer iter.Stop()

	var documents []BackupDocument

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate documents: %w", err)
		}

		backupDoc, err := s.backupDocument(ctx, doc)
		if err != nil {
			return nil, fmt.Errorf("failed to backup document %s: %w", doc.Ref.ID, err)
		}

		documents = append(documents, *backupDoc)
	}

	return &BackupData{
		Collection: collectionName,
		Documents:  documents,
		BackedUpAt: time.Now(),
	}, nil
}

// backupDocument backs up a document and all its subcollections
func (s *BackupService) backupDocument(ctx context.Context, doc *firestore.DocumentSnapshot) (*BackupDocument, error) {
	backupDoc := &BackupDocument{
		ID:             doc.Ref.ID,
		Data:           doc.Data(),
		SubCollections: make(map[string]BackupData),
	}

	// Get all subcollections
	subCollections, err := doc.Ref.Collections(ctx).GetAll()
	if err != nil {
		// If there are no subcollections, this is fine
		return backupDoc, nil
	}

	// Backup each subcollection
	for _, subCol := range subCollections {
		subColData, err := s.backupSubCollection(ctx, subCol)
		if err != nil {
			return nil, fmt.Errorf("failed to backup subcollection %s: %w", subCol.ID, err)
		}
		backupDoc.SubCollections[subCol.ID] = *subColData
	}

	return backupDoc, nil
}

// backupSubCollection backs up a subcollection
func (s *BackupService) backupSubCollection(ctx context.Context, subCol *firestore.CollectionRef) (*BackupData, error) {
	iter := subCol.Documents(ctx)
	defer iter.Stop()

	var documents []BackupDocument

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate subcollection documents: %w", err)
		}

		backupDoc, err := s.backupDocument(ctx, doc)
		if err != nil {
			return nil, fmt.Errorf("failed to backup subcollection document %s: %w", doc.Ref.ID, err)
		}

		documents = append(documents, *backupDoc)
	}

	return &BackupData{
		Collection: subCol.ID,
		Documents:  documents,
		BackedUpAt: time.Now(),
	}, nil
}
