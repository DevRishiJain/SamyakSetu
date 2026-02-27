package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB holds the database client and a reference to the main database.
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// Connect establishes a connection to MongoDB and returns a MongoDB instance.
func Connect(uri string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	// Verify the connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database("samyaksetu")
	log.Println("INFO: Connected to MongoDB — database: samyaksetu")

	mdb := &MongoDB{
		Client:   client,
		Database: db,
	}

	mdb.ensureIndexes(ctx)

	return mdb, nil
}

// Disconnect gracefully closes the MongoDB connection.
func (m *MongoDB) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.Client.Disconnect(ctx); err != nil {
		log.Printf("ERROR: MongoDB disconnect failed: %v", err)
	} else {
		log.Println("INFO: MongoDB connection closed")
	}
}

// Collection returns a handle to a named collection.
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

// ensureIndexes creates indexes for optimal query performance.
func (m *MongoDB) ensureIndexes(ctx context.Context) {
	// Unique index on farmer phone number
	farmersCol := m.Database.Collection("farmers")
	_, err := farmersCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "phone", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("WARN: Failed to create phone index: %v", err)
	}

	// Index on soil_data.farmerId for quick lookups
	soilCol := m.Database.Collection("soil_data")
	_, err = soilCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "farmerId", Value: 1}},
	})
	if err != nil {
		log.Printf("WARN: Failed to create soil farmerId index: %v", err)
	}

	// Index on chat_messages.farmerId + createdAt for sorted history
	chatCol := m.Database.Collection("chat_messages")
	_, err = chatCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "farmerId", Value: 1},
			{Key: "createdAt", Value: -1},
		},
	})
	if err != nil {
		log.Printf("WARN: Failed to create chat index: %v", err)
	}

	log.Println("INFO: Database indexes ensured")
}
