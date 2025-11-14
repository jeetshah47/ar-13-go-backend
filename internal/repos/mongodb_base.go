package repos

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoBaseRepo provides common MongoDB repository functionality
type MongoBaseRepo struct {
	client         *mongo.Client
	database       *mongo.Database
	collection     *mongo.Collection
	collectionName string
}

// NewMongoBaseRepo creates a new MongoDB base repository
func NewMongoBaseRepo(client *mongo.Client, dbName, collectionName string) *MongoBaseRepo {
	db := client.Database(dbName)
	collection := db.Collection(collectionName)

	return &MongoBaseRepo{
		client:         client,
		database:       db,
		collection:     collection,
		collectionName: collectionName,
	}
}

// GetCollection returns the MongoDB collection
func (r *MongoBaseRepo) GetCollection() *mongo.Collection {
	return r.collection
}

// GetByID gets an item by ID
func (r *MongoBaseRepo) GetByID(ctx context.Context, id string) *mongo.SingleResult {
	// Check both root level "id" and nested "model.id" (for embedded structs)
	filter := bson.M{
		"$or": []bson.M{
			{"id": id},
			{"model.id": id},
		},
	}
	return r.collection.FindOne(ctx, filter)
}

// Exists checks if an item exists
func (r *MongoBaseRepo) Exists(ctx context.Context, id string) (bool, error) {
	// Check both root level "id" and nested "model.id" (for embedded structs)
	filter := bson.M{
		"$or": []bson.M{
			{"id": id},
			{"model.id": id},
		},
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteByID deletes an item by ID
func (r *MongoBaseRepo) DeleteByID(ctx context.Context, id string) error {
	// Check both root level "id" and nested "model.id" (for embedded structs)
	filter := bson.M{
		"$or": []bson.M{
			{"id": id},
			{"model.id": id},
		},
	}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// InsertOne inserts a single document
func (r *MongoBaseRepo) InsertOne(ctx context.Context, document interface{}) error {
	_, err := r.collection.InsertOne(ctx, document)
	return err
}

// UpdateOne updates a single document
func (r *MongoBaseRepo) UpdateOne(ctx context.Context, id string, update bson.M) error {
	// Check both root level "id" and nested "model.id" (for embedded structs)
	filter := bson.M{
		"$or": []bson.M{
			{"id": id},
			{"model.id": id},
		},
	}

	// Add updatedAt timestamp if not already present
	if update["updatedAt"] == nil {
		update["updatedAt"] = time.Now()
	}

	updateDoc := bson.M{"$set": update}
	_, err := r.collection.UpdateOne(ctx, filter, updateDoc)
	return err
}

// FindAll finds all documents with optional filter
func (r *MongoBaseRepo) FindAll(ctx context.Context, filter bson.M, limit *int64, sort ...bson.M) ([]map[string]interface{}, error) {
	opts := options.Find()
	if limit != nil {
		opts.SetLimit(*limit)
	}

	// Default sort by creation date (descending) if no sort specified
	if len(sort) > 0 {
		opts.SetSort(sort[0])
	} else {
		opts.SetSort(bson.M{"created": -1}) // Try "created" first
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// FindOne finds a single document
func (r *MongoBaseRepo) FindOne(ctx context.Context, filter bson.M) *mongo.SingleResult {
	return r.collection.FindOne(ctx, filter)
}

// QueryByField queries documents by a specific field (replaces GSI queries)
func (r *MongoBaseRepo) QueryByField(ctx context.Context, fieldName, fieldValue string) ([]map[string]interface{}, error) {
	filter := bson.M{fieldName: fieldValue}
	return r.FindAll(ctx, filter, nil)
}

// QueryByFieldWithLimit queries documents by a specific field with limit
func (r *MongoBaseRepo) QueryByFieldWithLimit(ctx context.Context, fieldName, fieldValue string, limit int64) ([]map[string]interface{}, error) {
	filter := bson.M{fieldName: fieldValue}
	return r.FindAll(ctx, filter, &limit)
}

// BatchGetItems retrieves multiple items by IDs
func (r *MongoBaseRepo) BatchGetItems(ctx context.Context, ids []string) ([]map[string]interface{}, error) {
	if len(ids) == 0 {
		return []map[string]interface{}{}, nil
	}

	filter := bson.M{"id": bson.M{"$in": ids}}
	return r.FindAll(ctx, filter, nil)
}

// BatchInsertMany inserts multiple documents
func (r *MongoBaseRepo) BatchInsertMany(ctx context.Context, documents []interface{}) error {
	if len(documents) == 0 {
		return nil
	}

	_, err := r.collection.InsertMany(ctx, documents)
	return err
}

// CountDocuments counts documents matching a filter
func (r *MongoBaseRepo) CountDocuments(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

// UpdateMany updates multiple documents
func (r *MongoBaseRepo) UpdateMany(ctx context.Context, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	updateDoc := bson.M{"$set": update}
	return r.collection.UpdateMany(ctx, filter, updateDoc)
}

// DeleteMany deletes multiple documents
func (r *MongoBaseRepo) DeleteMany(ctx context.Context, filter bson.M) (*mongo.DeleteResult, error) {
	return r.collection.DeleteMany(ctx, filter)
}

// Aggregate performs an aggregation pipeline
func (r *MongoBaseRepo) Aggregate(ctx context.Context, pipeline []bson.M) (*mongo.Cursor, error) {
	return r.collection.Aggregate(ctx, pipeline)
}

// CreateIndex creates an index on the collection
func (r *MongoBaseRepo) CreateIndex(ctx context.Context, keys bson.M, unique bool) (string, error) {
	opts := options.Index().SetUnique(unique)
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: opts,
	}
	return r.collection.Indexes().CreateOne(ctx, indexModel)
}

// CreateCompoundIndex creates a compound index
func (r *MongoBaseRepo) CreateCompoundIndex(ctx context.Context, keys bson.M, unique bool) (string, error) {
	opts := options.Index().SetUnique(unique)
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: opts,
	}
	return r.collection.Indexes().CreateOne(ctx, indexModel)
}

// UnmarshalBSON unmarshals a MongoDB document into a struct
func UnmarshalBSON(data []byte, target interface{}) error {
	return bson.Unmarshal(data, target)
}

// MarshalBSON marshals a struct into BSON
func MarshalBSON(item interface{}) ([]byte, error) {
	return bson.Marshal(item)
}

// ConvertToObjectID converts a string to ObjectID
func ConvertToObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}

// IsObjectIDValid checks if a string is a valid ObjectID
func IsObjectIDValid(id string) bool {
	_, err := primitive.ObjectIDFromHex(id)
	return err == nil
}
