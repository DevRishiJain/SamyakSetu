// All rights reserved Samyak-Setu

package services

import (
	"crypto/rand"
	"io"
	"log"
)

// OTPService handles sending SMS OTPs to a phone number.
type OTPService interface {
	SendOTP(phone, code string) error
}

// MockOTPService logs the OTP to the console instead of sending a real SMS.
// Useful for development and testing without incurring API costs.
type MockOTPService struct{}

func NewMockOTPService() *MockOTPService {
	return &MockOTPService{}
}

func (s *MockOTPService) SendOTP(phone, code string) error {
	log.Printf("---------------------------------------------------------")
	log.Printf("[MOCK SMS] To: %s | Message: Your SamyakSetu verification code is: %s", phone, code)
	log.Printf("---------------------------------------------------------")
	return nil
}

// GenerateOTP generates a random 6-digit number string.
func GenerateOTP() string {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

	b := make([]byte, 6)
	n, err := io.ReadAtLeast(rand.Reader, b, 6)
	if n != 6 {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}
