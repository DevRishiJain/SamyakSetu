package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/models"
	"github.com/samyaksetu/backend/repositories"
	"github.com/samyaksetu/backend/services"
	"github.com/samyaksetu/backend/utils"
)

// AuthController handles authentication-related requests.
type AuthController struct {
	otpRepo    *repositories.OTPRepository
	otpService services.OTPService
}

// NewAuthController creates a new AuthController instance.
func NewAuthController(otpRepo *repositories.OTPRepository, otpService services.OTPService) *AuthController {
	return &AuthController{
		otpRepo:    otpRepo,
		otpService: otpService,
	}
}

// SendOTP handles POST /api/auth/send-otp — generates and sends an OTP via SMS.
func (ac *AuthController) SendOTP(c *gin.Context) {
	var req models.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Validate phone
	if err := utils.ValidatePhone(req.Phone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate 6-digit OTP
	code := services.GenerateOTP()

	// Save to database (expires in 5 minutes)
	if err := ac.otpRepo.SaveOTP(req.Phone, code); err != nil {
		log.Printf("ERROR: Failed to save OTP to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process OTP request"})
		return
	}

	// Send via SMS Service
	if err := ac.otpService.SendOTP(req.Phone, code); err != nil {
		log.Printf("ERROR: Failed to send SMS for %s: %v", req.Phone, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send SMS OTP"})
		return
	}

	log.Printf("INFO: OTP successfully sent to %s", req.Phone)
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully. It will expire in 5 minutes."})
}
