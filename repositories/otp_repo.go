package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/samyaksetu/backend/database"
	"github.com/samyaksetu/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// OTPRepository handles database operations for OTP verification.
type OTPRepository struct {
	db *database.MongoDB
}

// NewOTPRepository creates a new OTP repository and ensures index for expiry.
func NewOTPRepository(db *database.MongoDB) *OTPRepository {
	return &OTPRepository{
		db: db,
	}
}

// SaveOTP stores the OTP code with a 5-minute expiration time.
func (r *OTPRepository) SaveOTP(phone, code string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Delete any existing OTP for this phone number to avoid duplicates
	_, _ = r.db.Collection("otp_codes").DeleteMany(ctx, bson.M{"phone": phone})

	otp := models.OTP{
		Phone:     phone,
		Code:      code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	_, err := r.db.Collection("otp_codes").InsertOne(ctx, otp)
	return err
}

// VerifyOTP checks if the supplied code is valid and unexpired for the phone number.
func (r *OTPRepository) VerifyOTP(phone, code string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var otp models.OTP
	err := r.db.Collection("otp_codes").FindOne(ctx, bson.M{
		"phone":     phone,
		"code":      code,
		"expiresAt": bson.M{"$gt": time.Now()}, // Ensure not expired
	}).Decode(&otp)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("invalid or expired OTP")
		}
		return err
	}

	// Delete OTP after successful verification so it can't be reused
	_, _ = r.db.Collection("otp_codes").DeleteOne(ctx, bson.M{"_id": otp.ID})

	return nil
}
