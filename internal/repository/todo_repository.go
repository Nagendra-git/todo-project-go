// Package repository abstracts data persistence so handlers don't
// depend directly on MongoDB's driver/query syntax.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/example/todo-go-app/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrNotFound is returned when a requested todo does not exist.
var ErrNotFound = errors.New("todo not found")

// TodoRepository defines the persistence operations available for todos.
// Handlers depend on this interface, not on Mongo directly — which also
// makes it straightforward to swap in a fake for unit tests.
type TodoRepository interface {
	Create(ctx context.Context, todo *models.Todo) error
	List(ctx context.Context) ([]models.Todo, error)
	Get(ctx context.Context, id primitive.ObjectID) (*models.Todo, error)
	Update(ctx context.Context, id primitive.ObjectID, update models.TodoUpdate) (*models.Todo, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type mongoTodoRepository struct {
	collection *mongo.Collection
}

// NewMongoTodoRepository builds a TodoRepository backed by the given collection.
func NewMongoTodoRepository(collection *mongo.Collection) TodoRepository {
	return &mongoTodoRepository{collection: collection}
}

func (r *mongoTodoRepository) Create(ctx context.Context, todo *models.Todo) error {
	todo.ID = primitive.NewObjectID()
	now := time.Now()
	todo.CreatedAt = now
	todo.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, todo)
	return err
}

func (r *mongoTodoRepository) List(ctx context.Context) ([]models.Todo, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var todos []models.Todo
	if err := cursor.All(ctx, &todos); err != nil {
		return nil, err
	}
	if todos == nil {
		todos = []models.Todo{}
	}
	return todos, nil
}

func (r *mongoTodoRepository) Get(ctx context.Context, id primitive.ObjectID) (*models.Todo, error) {
	var todo models.Todo
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&todo)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *mongoTodoRepository) Update(ctx context.Context, id primitive.ObjectID, update models.TodoUpdate) (*models.Todo, error) {
	setFields := bson.M{"updated_at": time.Now()}
	if update.Title != nil {
		setFields["title"] = *update.Title
	}
	if update.Done != nil {
		setFields["done"] = *update.Done
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": setFields})
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, ErrNotFound
	}

	return r.Get(ctx, id)
}

func (r *mongoTodoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
