package repos

import (
	"context"

	"github.com/ar-13-go-backend/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

// TaskStatusRepo handles task status data operations with MongoDB
type TaskStatusRepo struct {
	*MongoBaseRepo
}

// NewTaskStatusRepo creates a new MongoDB task status repository
func NewTaskStatusRepo() *TaskStatusRepo {
	client := mongodb.GetClient()
	if client == nil {
		panic("MongoDB client is not initialized. Please ensure MongoDB is connected before creating repositories.")
	}
	return &TaskStatusRepo{
		MongoBaseRepo: NewMongoBaseRepo(client, mongodb.GetDatabaseName(), "task_statuses"),
	}
}

// GetAll gets all task statuses ordered by the order field
func (r *TaskStatusRepo) GetAll(ctx context.Context) ([]map[string]interface{}, error) {
	filter := bson.M{}
	// Sort by order field ascending
	return r.FindAll(ctx, filter, nil, bson.M{"order": 1})
}

