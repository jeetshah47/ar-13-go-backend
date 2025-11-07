package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/ar-13-go-backend/internal/models"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// VacationRepo handles vacation/leave request data operations
type VacationRepo struct {
	*DynamoBaseRepo
}

// NewVacationRepo creates a new vacation repository
func NewVacationRepo() *VacationRepo {
	return &VacationRepo{
		DynamoBaseRepo: NewDynamoBaseRepo("leaveRequests"),
	}
}

// GetByID gets a leave request by ID
func (r *VacationRepo) GetByID(ctx context.Context, id string) (*models.LeaveRequest, error) {
	item, err := r.DynamoBaseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	var request models.LeaveRequest
	if err := UnmarshalItem(item, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal leave request: %w", err)
	}

	return &request, nil
}

// GetByUserID gets all leave requests for a user
func (r *VacationRepo) GetByUserID(ctx context.Context, userID string) ([]models.LeaveRequest, error) {
	items, err := r.QueryByIndex(ctx, "userId-index", "userId", userID)
	if err != nil {
		return nil, err
	}

	requests := make([]models.LeaveRequest, 0, len(items))
	for _, item := range items {
		var request models.LeaveRequest
		if err := UnmarshalItem(item, &request); err != nil {
			return nil, fmt.Errorf("failed to unmarshal leave request: %w", err)
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// GetAll gets all leave requests
func (r *VacationRepo) GetAll(ctx context.Context) ([]models.LeaveRequest, error) {
	items, err := r.ScanItems(ctx, nil)
	if err != nil {
		return nil, err
	}

	requests := make([]models.LeaveRequest, 0, len(items))
	for _, item := range items {
		var request models.LeaveRequest
		if err := UnmarshalItem(item, &request); err != nil {
			return nil, fmt.Errorf("failed to unmarshal leave request: %w", err)
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
	items, err := r.QueryByIndex(ctx, "status-index", "status", string(status))
	if err != nil {
		return nil, err
	}

	requests := make([]models.LeaveRequest, 0, len(items))
	for _, item := range items {
		var request models.LeaveRequest
		if err := UnmarshalItem(item, &request); err != nil {
			return nil, fmt.Errorf("failed to unmarshal leave request: %w", err)
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// GetByType gets leave requests by type
func (r *VacationRepo) GetByType(ctx context.Context, requestType models.LeaveRequestType) ([]models.LeaveRequest, error) {
	items, err := r.ScanItems(ctx, nil)
	if err != nil {
		return nil, err
	}

	requests := make([]models.LeaveRequest, 0)
	for _, item := range items {
		// Filter by requestType
		if rt, ok := item["requestType"].(*types.AttributeValueMemberS); ok && rt.Value == string(requestType) {
			var request models.LeaveRequest
			if err := UnmarshalItem(item, &request); err != nil {
				return nil, fmt.Errorf("failed to unmarshal leave request: %w", err)
			}
			requests = append(requests, request)
		}
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

	data := map[string]interface{}{
		"id":           request.ID,
		"userId":       request.UserID,
		"requestType":  string(request.RequestType),
		"startDate":    request.StartDate.Format(time.RFC3339),
		"duration":     request.Duration,
		"durationType": string(request.DurationType),
		"status":       string(request.Status),
		"requestedAt":  request.RequestedAt.Format(time.RFC3339),
		"created":      request.Created.Format(time.RFC3339),
	}

	if request.EndDate != nil {
		data["endDate"] = request.EndDate.Format(time.RFC3339)
	}
	if request.Comments != nil {
		data["comments"] = *request.Comments
	}
	if request.ReviewedBy != nil {
		data["reviewedBy"] = *request.ReviewedBy
	}
	if request.ReviewedAt != nil {
		data["reviewedAt"] = request.ReviewedAt.Format(time.RFC3339)
	}
	if request.ReviewComments != nil {
		data["reviewComments"] = *request.ReviewComments
	}
	if request.WorkingHours != nil {
		data["workingHours"] = map[string]interface{}{
			"from": request.WorkingHours.From,
			"to":   request.WorkingHours.To,
		}
	}

	return r.PutItem(ctx, data)
}

// Update updates a leave request
func (r *VacationRepo) Update(ctx context.Context, request *models.LeaveRequest) error {
	now := time.Now()
	request.Updated = &now

	updates := map[string]interface{}{
		"userId":       request.UserID,
		"requestType":  string(request.RequestType),
		"startDate":    request.StartDate.Format(time.RFC3339),
		"duration":     request.Duration,
		"durationType": string(request.DurationType),
		"status":       string(request.Status),
		"updated":      request.Updated.Format(time.RFC3339),
	}

	if request.EndDate != nil {
		updates["endDate"] = request.EndDate.Format(time.RFC3339)
	}
	if request.Comments != nil {
		updates["comments"] = *request.Comments
	}
	if request.ReviewedBy != nil {
		updates["reviewedBy"] = *request.ReviewedBy
	}
	if request.ReviewedAt != nil {
		updates["reviewedAt"] = request.ReviewedAt.Format(time.RFC3339)
	}
	if request.ReviewComments != nil {
		updates["reviewComments"] = *request.ReviewComments
	}
	if request.WorkingHours != nil {
		updates["workingHours"] = map[string]interface{}{
			"from": request.WorkingHours.From,
			"to":   request.WorkingHours.To,
		}
	}

	return r.UpdateItem(ctx, request.ID, updates)
}

// UpdateStatus updates leave request status
func (r *VacationRepo) UpdateStatus(ctx context.Context, id string, status models.LeaveRequestStatus, reviewedBy string, reviewComments *string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     string(status),
		"reviewedBy": reviewedBy,
		"reviewedAt": now.Format(time.RFC3339),
	}

	if reviewComments != nil {
		updates["reviewComments"] = *reviewComments
	}

	return r.UpdateItem(ctx, id, updates)
}

// Delete deletes a leave request
func (r *VacationRepo) Delete(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
