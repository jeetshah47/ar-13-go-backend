package repos

import (
	"context"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/ar-13-go-backend/pkg/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// VacationRepo handles vacation/leave request data operations with MongoDB
type VacationRepo struct {
	*MongoBaseRepo
}

// NewVacationRepo creates a new MongoDB vacation repository
func NewVacationRepo() *VacationRepo {
	client := mongodb.GetClient()
	dbName := mongodb.GetDatabaseName()
	return &VacationRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, dbName, "leaveRequests"),
	}
}

// GetByID gets a leave request by ID
func (r *VacationRepo) GetByID(ctx context.Context, id string) (*models.LeaveRequest, error) {
	result := r.MongoBaseRepo.GetByID(ctx, id)

	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var request models.LeaveRequest
	if err := result.Decode(&request); err != nil {
		return nil, err
	}

	return &request, nil
}

// GetByUserID gets all leave requests for a user
func (r *VacationRepo) GetByUserID(ctx context.Context, userID string) ([]models.LeaveRequest, error) {
	filter := bson.M{"userId": userID}
	items, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	requests := make([]models.LeaveRequest, 0, len(items))
	for _, item := range items {
		var request models.LeaveRequest
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &request); err != nil {
			continue
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// GetAll gets all leave requests
func (r *VacationRepo) GetAll(ctx context.Context) ([]models.LeaveRequest, error) {
	items, err := r.FindAll(ctx, bson.M{}, nil)
	if err != nil {
		return nil, err
	}

	requests := make([]models.LeaveRequest, 0, len(items))
	for _, item := range items {
		var request models.LeaveRequest
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &request); err != nil {
			continue
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// GetPending gets all pending leave requests
func (r *VacationRepo) GetPending(ctx context.Context) ([]models.LeaveRequest, error) {
	return r.GetByStatus(ctx, models.LeaveRequestStatusPending)
}

// GetByStatus gets leave requests by status
func (r *VacationRepo) GetByStatus(ctx context.Context, status models.LeaveRequestStatus) ([]models.LeaveRequest, error) {
	filter := bson.M{"status": string(status)}
	items, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	requests := make([]models.LeaveRequest, 0, len(items))
	for _, item := range items {
		var request models.LeaveRequest
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &request); err != nil {
			continue
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// GetByType gets leave requests by type
func (r *VacationRepo) GetByType(ctx context.Context, requestType models.LeaveRequestType) ([]models.LeaveRequest, error) {
	filter := bson.M{"requestType": string(requestType)}
	items, err := r.FindAll(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	requests := make([]models.LeaveRequest, 0, len(items))
	for _, item := range items {
		var request models.LeaveRequest
		bsonBytes, _ := bson.Marshal(item)
		if err := bson.Unmarshal(bsonBytes, &request); err != nil {
			continue
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// Add creates a new leave request
func (r *VacationRepo) Add(ctx context.Context, request *models.LeaveRequest) error {
	now := time.Now()
	if request.ID == "" {
		request.ID = uuid.New().String()
	}
	request.Created = now
	request.RequestedAt = now

	return r.InsertOne(ctx, request)
}

// Update updates a leave request
func (r *VacationRepo) Update(ctx context.Context, request *models.LeaveRequest) error {
	now := time.Now()
	request.Updated = &now

	updates := bson.M{
		"userId":       request.UserID,
		"requestType":  string(request.RequestType),
		"startDate":    request.StartDate,
		"duration":     request.Duration,
		"durationType": string(request.DurationType),
		"status":       string(request.Status),
		"updated":      request.Updated,
	}

	if request.EndDate != nil {
		updates["endDate"] = request.EndDate
	}
	if request.Comments != nil {
		updates["comments"] = *request.Comments
	}
	if request.ReviewedBy != nil {
		updates["reviewedBy"] = *request.ReviewedBy
	}
	if request.ReviewedAt != nil {
		updates["reviewedAt"] = request.ReviewedAt
	}
	if request.ReviewComments != nil {
		updates["reviewComments"] = *request.ReviewComments
	}
	if request.WorkingHours != nil {
		updates["workingHours"] = request.WorkingHours
	}

	return r.UpdateOne(ctx, request.ID, updates)
}

// UpdateStatus updates leave request status
func (r *VacationRepo) UpdateStatus(ctx context.Context, id string, status models.LeaveRequestStatus, reviewedBy string, reviewComments *string) error {
	now := time.Now()
	updates := bson.M{
		"status":     string(status),
		"reviewedBy": reviewedBy,
		"reviewedAt": now,
	}

	if reviewComments != nil {
		updates["reviewComments"] = *reviewComments
	}

	return r.UpdateOne(ctx, id, updates)
}

// Delete deletes a leave request
func (r *VacationRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}

