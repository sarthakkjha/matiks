// Package database provides MongoDB connection management.
// Handles connection, disconnection, and collection access.
package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client   *mongo.Client
	database *mongo.Database
)

// Connect establishes a connection to MongoDB.
// Uses the provided URI with sensible timeout defaults.
// Returns an error if connection fails.
func Connect(ctx context.Context, uri string) error {
	var err error

	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(30 * time.Second).
		SetServerSelectionTimeout(30 * time.Second)

	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	if err = client.Ping(ctx, nil); err != nil {
		return err
	}

	database = client.Database("matiks-leaderboard")
	log.Println("✅ MongoDB connected successfully")
	return nil
}

// Disconnect closes the MongoDB connection.
// Should be called when the application shuts down.
func Disconnect(ctx context.Context) {
	if client != nil {
		if err := client.Disconnect(ctx); err != nil {
			log.Println("Error disconnecting from MongoDB:", err)
		}
	}
}

// Collection returns a MongoDB collection by name.
func Collection(name string) *mongo.Collection {
	return database.Collection(name)
}

// DB returns the database instance.
func DB() *mongo.Database {
	return database
}
