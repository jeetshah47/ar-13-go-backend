package repos

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ar-13-go-backend/internal/models"
	"google.golang.org/api/iterator"
)

// VacationRepo handles vacation/leave request data operations
type VacationRepo struct {
	*BaseRepo
}

// NewVacationRepo creates a new vacation repository
func NewVacationRepo() *VacationRepo {
	return &VacationRepo{
		BaseRepo: NewBaseRepo("leaveRequests"),
	}
}

// GetByID gets a leave request by ID
func (r *VacationRepo) GetByID(ctx context.Context, id string) (*models.LeaveRequest, error) {
	doc, err := r.collection.Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	if !doc.Exists() {
		return nil, nil
	}

	data := doc.Data()
	// Convert time fields from strings/timestamps to time.Time
	if err := ConvertTimeFieldsInMap(data, []string{"startDate", "endDate", "requestedAt", "reviewedAt", "created", "updated"}); err != nil {
		return nil, err
	}

	var request models.LeaveRequest
	if err := doc.DataTo(&request); err != nil {
		return nil, err
	}
	request.ID = doc.Ref.ID
	return &request, nil
}

// GetByUserID gets all leave requests for a user
func (r *VacationRepo) GetByUserID(ctx context.Context, userID string) ([]models.LeaveRequest, error) {
	iter := r.collection.Where("userId", "==", userID).Documents(ctx)
	var requests []models.LeaveRequest

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		// Convert time fields from strings/timestamps to time.Time
		if err := ConvertTimeFieldsInMap(data, []string{"startDate", "endDate", "requestedAt", "reviewedAt", "created", "updated"}); err != nil {
			return nil, err
		}

		var request models.LeaveRequest
		if err := doc.DataTo(&request); err != nil {
			return nil, err
		}
		request.ID = doc.Ref.ID
		requests = append(requests, request)
	}

	return requests, nil
}

// GetAll gets all leave requests
func (r *VacationRepo) GetAll(ctx context.Context) ([]models.LeaveRequest, error) {
	iter := r.collection.Documents(ctx)
	var requests []models.LeaveRequest

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		// Convert time fields from strings/timestamps to time.Time
		if err := ConvertTimeFieldsInMap(data, []string{"startDate", "endDate", "requestedAt", "reviewedAt", "created", "updated"}); err != nil {
			return nil, err
		}

		var request models.LeaveRequest
		if err := doc.DataTo(&request); err != nil {
			return nil, err
		}
		request.ID = doc.Ref.ID
		requests = append(requests, request)
	}

	return requests, nil
}

// GetPending gets all pending leave requests
func (r *VacationRepo) GetPending(ctx context.Context) ([]models.LeaveRequest, error) {
	iter := r.collection.Where("status", "==", string(models.LeaveRequestStatusPending)).Documents(ctx)
	var requests []models.LeaveRequest

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		// Convert time fields from strings/timestamps to time.Time
		if err := ConvertTimeFieldsInMap(data, []string{"startDate", "endDate", "requestedAt", "reviewedAt", "created", "updated"}); err != nil {
			return nil, err
		}

		var request models.LeaveRequest
		if err := doc.DataTo(&request); err != nil {
			return nil, err
		}
		request.ID = doc.Ref.ID
		requests = append(requests, request)
	}

	return requests, nil
}

// GetByStatus gets leave requests by status
func (r *VacationRepo) GetByStatus(ctx context.Context, status models.LeaveRequestStatus) ([]models.LeaveRequest, error) {
	iter := r.collection.Where("status", "==", string(status)).Documents(ctx)
	var requests []models.LeaveRequest

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		// Convert time fields from strings/timestamps to time.Time
		if err := ConvertTimeFieldsInMap(data, []string{"startDate", "endDate", "requestedAt", "reviewedAt", "created", "updated"}); err != nil {
			return nil, err
		}

		var request models.LeaveRequest
		if err := doc.DataTo(&request); err != nil {
			return nil, err
		}
		request.ID = doc.Ref.ID
		requests = append(requests, request)
	}

	return requests, nil
}

// GetByType gets leave requests by type
func (r *VacationRepo) GetByType(ctx context.Context, requestType models.LeaveRequestType) ([]models.LeaveRequest, error) {
	iter := r.collection.Where("requestType", "==", string(requestType)).Documents(ctx)
	var requests []models.LeaveRequest

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		// Convert time fields from strings/timestamps to time.Time
		if err := ConvertTimeFieldsInMap(data, []string{"startDate", "endDate", "requestedAt", "reviewedAt", "created", "updated"}); err != nil {
			return nil, err
		}

		var request models.LeaveRequest
		if err := doc.DataTo(&request); err != nil {
			return nil, err
		}
		request.ID = doc.Ref.ID
		requests = append(requests, request)
	}

	return requests, nil
}

// Add creates a new leave request
func (r *VacationRepo) Add(ctx context.Context, request *models.LeaveRequest) error {
	newDocRef := r.collection.NewDoc()
	request.ID = newDocRef.ID
	request.Created = time.Now()
	request.RequestedAt = time.Now()

	data := map[string]interface{}{
		"id":           request.ID,
		"userId":       request.UserID,
		"requestType":  string(request.RequestType),
		"startDate":    request.StartDate,
		"duration":     request.Duration,
		"durationType": string(request.DurationType),
		"status":       string(request.Status),
		"requestedAt":  request.RequestedAt,
		"created":      request.Created,
	}
	if request.EndDate != nil {
		data["endDate"] = *request.EndDate
	}
	if request.Comments != nil {
		data["comments"] = *request.Comments
	}
	if request.ReviewedBy != nil {
		data["reviewedBy"] = *request.ReviewedBy
	}
	if request.ReviewedAt != nil {
		data["reviewedAt"] = *request.ReviewedAt
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

	_, err := newDocRef.Set(ctx, data)
	return err
}

// Update updates a leave request
func (r *VacationRepo) Update(ctx context.Context, request *models.LeaveRequest) error {
	now := time.Now()
	request.Updated = &now

	data := map[string]interface{}{
		"id":           request.ID,
		"userId":       request.UserID,
		"requestType":  string(request.RequestType),
		"startDate":    request.StartDate,
		"duration":     request.Duration,
		"durationType": string(request.DurationType),
		"status":       string(request.Status),
		"updated":      now,
	}
	if request.EndDate != nil {
		data["endDate"] = *request.EndDate
	}
	if request.Comments != nil {
		data["comments"] = *request.Comments
	}
	if request.ReviewedBy != nil {
		data["reviewedBy"] = *request.ReviewedBy
	}
	if request.ReviewedAt != nil {
		data["reviewedAt"] = *request.ReviewedAt
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

	_, err := r.collection.Doc(request.ID).Set(ctx, data)
	return err
}

// UpdateStatus updates leave request status
func (r *VacationRepo) UpdateStatus(ctx context.Context, id string, status models.LeaveRequestStatus, reviewedBy string, reviewComments *string) error {
	now := time.Now()
	updates := []firestore.Update{
		{Path: "status", Value: string(status)},
		{Path: "reviewedBy", Value: reviewedBy},
		{Path: "reviewedAt", Value: now},
		{Path: "updated", Value: now},
	}
	if reviewComments != nil {
		updates = append(updates, firestore.Update{Path: "reviewComments", Value: *reviewComments})
	}

	_, err := r.collection.Doc(id).Update(ctx, updates)
	return err
}

// Delete deletes a leave request
func (r *VacationRepo) Delete(ctx context.Context, id string) error {
	_, err := r.collection.Doc(id).Delete(ctx)
	return err
}
