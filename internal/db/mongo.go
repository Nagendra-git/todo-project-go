// Package db manages the MongoDB client lifecycle.
package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect opens a MongoDB connection and verifies it with a ping.
// Callers are responsible for calling Disconnect on the returned
// client during shutdown.
func Connect(uri string, timeout time.Duration) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect to mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	return client, nil
}

// Disconnect closes the MongoDB connection, logging but not failing
// hard if it errors during shutdown.
func Disconnect(client *mongo.Client, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return client.Disconnect(ctx)
}
