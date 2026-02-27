package repositories

import (
	"context"
	"time"

	"github.com/samyaksetu/backend/database"
	"github.com/samyaksetu/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SoilRepository handles all database operations for soil data.
type SoilRepository struct {
	db *database.MongoDB
}

// NewSoilRepository creates a new SoilRepository instance.
func NewSoilRepository(db *database.MongoDB) *SoilRepository {
	return &SoilRepository{db: db}
}

// Create inserts a new soil data record into the database.
func (r *SoilRepository) Create(soil *models.SoilData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	soil.CreatedAt = time.Now()
	result, err := r.db.Collection("soil_data").InsertOne(ctx, soil)
	if err != nil {
		return err
	}

	soil.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindLatestByFarmerID retrieves the most recent soil analysis for a farmer.
func (r *SoilRepository) FindLatestByFarmerID(farmerID primitive.ObjectID) (*models.SoilData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	var soil models.SoilData
	err := r.db.Collection("soil_data").FindOne(ctx, bson.M{"farmerId": farmerID}, opts).Decode(&soil)
	if err != nil {
		return nil, err
	}

	return &soil, nil
}
