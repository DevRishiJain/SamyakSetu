package repositories

import (
	"context"
	"time"

	"github.com/samyaksetu/backend/database"
	"github.com/samyaksetu/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FarmerRepository handles all database operations for farmers.
type FarmerRepository struct {
	db *database.MongoDB
}

// NewFarmerRepository creates a new FarmerRepository instance.
func NewFarmerRepository(db *database.MongoDB) *FarmerRepository {
	return &FarmerRepository{db: db}
}

// Create inserts a new farmer into the database.
func (r *FarmerRepository) Create(farmer *models.Farmer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	farmer.CreatedAt = time.Now()
	result, err := r.db.Collection("farmers").InsertOne(ctx, farmer)
	if err != nil {
		return err
	}

	farmer.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID retrieves a farmer by their ObjectID.
func (r *FarmerRepository) FindByID(id primitive.ObjectID) (*models.Farmer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var farmer models.Farmer
	err := r.db.Collection("farmers").FindOne(ctx, bson.M{"_id": id}).Decode(&farmer)
	if err != nil {
		return nil, err
	}

	return &farmer, nil
}

// FindByPhone retrieves a farmer by their phone number.
func (r *FarmerRepository) FindByPhone(phone string) (*models.Farmer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var farmer models.Farmer
	err := r.db.Collection("farmers").FindOne(ctx, bson.M{"phone": phone}).Decode(&farmer)
	if err != nil {
		return nil, err
	}

	return &farmer, nil
}

// UpdateLocation updates a farmer's GPS coordinates.
func (r *FarmerRepository) UpdateLocation(id primitive.ObjectID, lat, lng float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"location.latitude":  lat,
			"location.longitude": lng,
		},
	}

	_, err := r.db.Collection("farmers").UpdateByID(ctx, id, update)
	return err
}
