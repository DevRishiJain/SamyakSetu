// All rights reserved Samyak-Setu

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OTP represents an OTP code for a phone number.
type OTP struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Phone     string             `bson:"phone"`
	Code      string             `bson:"code"`
	ExpiresAt time.Time          `bson:"expiresAt"`
	CreatedAt time.Time          `bson:"createdAt"`
}

// SendOTPRequest is the input for requesting an OTP.
type SendOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
}
