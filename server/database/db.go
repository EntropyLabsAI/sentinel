// Package database contains the database operations for the Sentinel server.
// It uses a MongoDB database to store the data.
package database

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database interface {
	Connect() (*mongo.Client, error)
	Disconnect(client *mongo.Client)
	GetCollection(client *mongo.Client, collectionName string) *mongo.Collection
}

// Connect to the MongoDB database.
func Connect() (*mongo.Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		return nil, fmt.Errorf("database: failed to connect to MongoDB: %w", err)
	}

	return client, nil
}

func Disconnect(client *mongo.Client) error {
	err := client.Disconnect(context.Background())
	if err != nil {
		return fmt.Errorf("database: failed to disconnect from MongoDB: %w", err)
	}

	return nil
}

// Get a collection from the database.
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("sentinel").Collection(collectionName)
}
