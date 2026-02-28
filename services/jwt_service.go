// All rights reserved Samyak-Setu

package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles creation and validation of JSON Web Tokens for session management.
type JWTService struct {
	secretKey []byte
	issuer    string
}

// JWTClaims extends standard JWT claims with farmer-specific data.
type JWTClaims struct {
	FarmerID string `json:"farmerId"`
	Phone    string `json:"phone"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service with the given secret key.
func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secretKey: []byte(secret),
		issuer:    "SamyakSetu",
	}
}

// GenerateToken creates a signed JWT token for the given farmer that expires in 7 days.
func (s *JWTService) GenerateToken(farmerID, phone, name string) (string, error) {
	claims := JWTClaims{
		FarmerID: farmerID,
		Phone:    phone,
		Name:     name,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken parses and validates the JWT token string and returns the claims.
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
