// Package models defines the domain types shared across the application.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Todo represents a single task.
type Todo struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title     string             `json:"title" bson:"title"`
	Done      bool               `json:"done" bson:"done"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// TodoUpdate captures the optional fields a client may patch.
// Pointers distinguish "not provided" from "set to zero value".
type TodoUpdate struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}
