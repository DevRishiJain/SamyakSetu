// All rights reserved Samyak-Setu

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SoilData represents an analyzed soil sample from a farmer's land.
type SoilData struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FarmerID  primitive.ObjectID `json:"farmerId" bson:"farmerId"`
	ImagePath string             `json:"imagePath" bson:"imagePath"`
	SoilType  string             `json:"soilType" bson:"soilType"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

// SoilUploadResponse is returned after a successful soil analysis.
type SoilUploadResponse struct {
	SoilType  string `json:"soilType"`
	ImagePath string `json:"imagePath"`
}
