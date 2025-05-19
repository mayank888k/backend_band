package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	client          *mongo.Client
	db              *mongo.Database
	dbOnce          sync.Once
	collectionNames = map[string]string{
		"bookings":    "bookings",
		"employees":   "employees",
		"payments":    "payments",
		"admin_users": "admin_users",
	}
)

func init() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading .env file, using default environment variables")
	}
}

// GetDB returns a singleton MongoDB database connection
func GetDB() *mongo.Database {
	dbOnce.Do(func() {
		// Get MongoDB URI and DB name from environment variable or use default
		mongoURI := os.Getenv("MongoURI")
		if mongoURI == "" {
			mongoURI = "mongodb://localhost:27017"
			log.Println("Warning: MongoURI environment variable not set, using default:", mongoURI)
		} else {
			log.Println("Using MongoDB URI from environment")
		}

		dbName := os.Getenv("DBName")
		if dbName == "" {
			dbName = "booking"
			log.Println("Warning: DBName environment variable not set, using default:", dbName)
		} else {
			log.Println("Using database name from environment:", dbName)
		}

		// Set client options - explicitly not using ServerAPI version to maximize compatibility
		clientOptions := options.Client().ApplyURI(mongoURI)

		log.Println("Connecting to MongoDB...")

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Connect to MongoDB
		var err error
		client, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		// Ping the database with a separate context and increased timeout
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer pingCancel()

		log.Println("Pinging MongoDB server...")
		if err = client.Ping(pingCtx, readpref.Primary()); err != nil {
			log.Fatalf("Failed to ping MongoDB: %v", err)
		}

		// Get database instance
		db = client.Database(dbName)
		log.Printf("Connected to MongoDB database: %s", dbName)

		// Create indexes
		indexCtx, indexCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer indexCancel()

		log.Println("Creating indexes...")
		if err = createIndexesInternal(indexCtx, db); err != nil {
			log.Printf("Warning: Failed to create indexes: %v", err)
			// Not fatal, continue anyway
		}

		log.Println("MongoDB setup completed successfully")
	})

	return db
}

// GetCollection returns a MongoDB collection by name
func GetCollection(name string) *mongo.Collection {
	// Check if collection name exists in our mapping
	collName, exists := collectionNames[name]
	if !exists {
		log.Printf("Warning: Unknown collection name: %s, using name directly", name)
		collName = name
	}

	return GetDB().Collection(collName)
}

// createIndexesInternal creates necessary indexes for collections
// This function directly uses the database instance to avoid recursion
func createIndexesInternal(ctx context.Context, database *mongo.Database) error {
	// Bookings collection indexes
	bookingsColl := database.Collection(collectionNames["bookings"])
	_, err := bookingsColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{"booking_id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"phone", 1}},
		},
		{
			Keys: bson.D{{"event_date", 1}},
		},
	})
	if err != nil {
		return fmt.Errorf("error creating bookings indexes: %w", err)
	}

	// Employees collection indexes
	employeesColl := database.Collection(collectionNames["employees"])
	_, err = employeesColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{"username", 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return fmt.Errorf("error creating employees indexes: %w", err)
	}

	// Admin users collection indexes
	adminUsersColl := database.Collection(collectionNames["admin_users"])
	_, err = adminUsersColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{"username", 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return fmt.Errorf("error creating admin_users indexes: %w", err)
	}

	// Payments collection indexes
	paymentsColl := database.Collection(collectionNames["payments"])
	_, err = paymentsColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"employee_id", 1}},
		},
		{
			Keys: bson.D{{"date", 1}},
		},
	})
	if err != nil {
		return fmt.Errorf("error creating payments indexes: %w", err)
	}

	return nil
}

// createIndexes is kept for backward compatibility
func createIndexes(ctx context.Context) error {
	// This now just calls the internal function with the already initialized db
	return createIndexesInternal(ctx, db)
}

// Close closes the MongoDB connection
func Close() {
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		} else {
			log.Println("MongoDB connection closed successfully")
		}
	}
}
