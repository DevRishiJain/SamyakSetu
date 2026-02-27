package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Location represents GPS coordinates for a farmer's land.
type Location struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

// Farmer represents a registered farmer in the system.
type Farmer struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Phone     string             `json:"phone" bson:"phone"`
	Location  Location           `json:"location" bson:"location"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

// SignupRequest is the expected input for farmer registration.
type SignupRequest struct {
	Name      string  `json:"name" binding:"required"`
	Phone     string  `json:"phone" binding:"required"`
	OTP       string  `json:"otp" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// UpdateLocationRequest is the expected input for location updates.
type UpdateLocationRequest struct {
	FarmerID  string  `json:"farmerId" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}
