// All rights reserved Samyak-Setu

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatMessage represents a single message in a farmer's chat history.
type ChatMessage struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FarmerID  primitive.ObjectID `json:"farmerId" bson:"farmerId"`
	Role      string             `json:"role" bson:"role"` // "user" or "ai"
	Message   string             `json:"message" bson:"message"`
	ImagePath string             `json:"imagePath,omitempty" bson:"imagePath,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

// ChatRequest is the expected input for the advisory chat endpoint.
type ChatRequest struct {
	FarmerID string `json:"farmerId" form:"farmerId" binding:"required"`
	Message  string `json:"message" form:"message" binding:"required"`
}

// ChatResponse is returned after a successful AI advisory chat.
type ChatResponse struct {
	Reply string `json:"reply"`
}
