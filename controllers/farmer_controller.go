// All rights reserved Samyak-Setu

package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/models"
	"github.com/samyaksetu/backend/repositories"
	"github.com/samyaksetu/backend/services"
	"github.com/samyaksetu/backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FarmerController handles HTTP requests related to farmers.
type FarmerController struct {
	farmerRepo     *repositories.FarmerRepository
	otpRepo        *repositories.OTPRepository
	jwtService     *services.JWTService
	storageService services.StorageService
	prototypeMode  bool
}

// NewFarmerController creates a new FarmerController instance.
func NewFarmerController(farmerRepo *repositories.FarmerRepository, otpRepo *repositories.OTPRepository, jwtService *services.JWTService, storageService services.StorageService, prototypeMode bool) *FarmerController {
	return &FarmerController{
		farmerRepo:     farmerRepo,
		otpRepo:        otpRepo,
		jwtService:     jwtService,
		storageService: storageService,
		prototypeMode:  prototypeMode,
	}
}

// Signup handles POST /api/signup — registers a new farmer and returns a JWT token.
func (fc *FarmerController) Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Validate phone
	if err := utils.ValidatePhone(req.Phone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate coordinates
	if err := utils.ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if phone already registered
	existing, _ := fc.farmerRepo.FindByPhone(req.Phone)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Phone number already registered"})
		return
	}

	// In prototype mode, master OTP "000000" always works (no need to call send-otp first)
	if fc.prototypeMode && req.OTP == "000000" {
		log.Printf("INFO: Prototype mode — skipping OTP verification for %s", req.Phone)
	} else {
		// Normal OTP verification
		if err := fc.otpRepo.VerifyOTP(req.Phone, req.OTP); err != nil {
			log.Printf("WARN: OTP verification failed for %s: %v", req.Phone, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
			return
		}
	}

	farmer := &models.Farmer{
		Name:  req.Name,
		Phone: req.Phone,
		Location: models.Location{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		},
	}

	if err := fc.farmerRepo.Create(farmer); err != nil {
		log.Printf("ERROR: Failed to create farmer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create farmer"})
		return
	}

	// Generate JWT token for the newly registered farmer
	token, err := fc.jwtService.GenerateToken(farmer.ID.Hex(), farmer.Phone, farmer.Name)
	if err != nil {
		log.Printf("ERROR: Failed to generate JWT for farmer %s: %v", farmer.ID.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration successful but failed to generate session token"})
		return
	}

	log.Printf("INFO: Farmer registered — id=%s name=%s phone=%s", farmer.ID.Hex(), farmer.Name, farmer.Phone)
	c.JSON(http.StatusCreated, gin.H{
		"id":         farmer.ID.Hex(),
		"name":       farmer.Name,
		"phone":      farmer.Phone,
		"location":   farmer.Location,
		"profilePic": farmer.ProfilePic,
		"createdAt":  farmer.CreatedAt,
		"token":      token,
	})
}

// Login handles POST /api/login — authenticates an existing farmer and returns a JWT token.
func (fc *FarmerController) Login(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
		OTP   string `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Validate phone
	if err := utils.ValidatePhone(req.Phone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the farmer
	farmer, err := fc.farmerRepo.FindByPhone(req.Phone)
	if err != nil || farmer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No account found with this phone number"})
		return
	}

	// In prototype mode, master OTP "000000" always works
	if fc.prototypeMode && req.OTP == "000000" {
		log.Printf("INFO: Prototype mode — skipping OTP verification for login %s", req.Phone)
	} else {
		if err := fc.otpRepo.VerifyOTP(req.Phone, req.OTP); err != nil {
			log.Printf("WARN: Login OTP verification failed for %s: %v", req.Phone, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
			return
		}
	}

	// Generate JWT token
	token, err := fc.jwtService.GenerateToken(farmer.ID.Hex(), farmer.Phone, farmer.Name)
	if err != nil {
		log.Printf("ERROR: Failed to generate JWT for farmer %s: %v", farmer.ID.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session token"})
		return
	}

	log.Printf("INFO: Farmer logged in — id=%s name=%s phone=%s", farmer.ID.Hex(), farmer.Name, farmer.Phone)
	c.JSON(http.StatusOK, gin.H{
		"id":         farmer.ID.Hex(),
		"name":       farmer.Name,
		"phone":      farmer.Phone,
		"location":   farmer.Location,
		"profilePic": farmer.ProfilePic,
		"token":      token,
	})
}

// Logout handles POST /api/logout — for prototype, just confirms logout (frontend clears the token).
func (fc *FarmerController) Logout(c *gin.Context) {
	farmerName, _ := c.Get("farmerName")
	log.Printf("INFO: Farmer logged out — name=%v", farmerName)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// UpdateLocation handles PUT /api/location — updates a farmer's GPS coordinates.
func (fc *FarmerController) UpdateLocation(c *gin.Context) {
	var req models.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Validate ObjectID
	farmerID, err := primitive.ObjectIDFromHex(req.FarmerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID format"})
		return
	}

	// Validate coordinates
	if err := utils.ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure farmer exists
	farmer, err := fc.farmerRepo.FindByID(farmerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Farmer not found"})
		return
	}

	if err := fc.farmerRepo.UpdateLocation(farmerID, req.Latitude, req.Longitude); err != nil {
		log.Printf("ERROR: Failed to update location for farmer %s: %v", farmerID.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update location"})
		return
	}

	log.Printf("INFO: Location updated — farmer=%s lat=%.6f lng=%.6f", farmer.Name, req.Latitude, req.Longitude)
	c.JSON(http.StatusOK, gin.H{
		"message":  "Location updated successfully",
		"farmerId": farmerID.Hex(),
		"location": models.Location{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		},
	})
}

// UploadProfilePic handles PUT /api/profile-pic
// Accepts a file upload and updates the farmer's profile picture.
func (fc *FarmerController) UploadProfilePic(c *gin.Context) {
	// Let's get the farmerID from the context since it's a protected route
	farmerIDVal, exists := c.Get("farmerId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Farmer ID not found in context"})
		return
	}
	farmerIDStr := farmerIDVal.(string)

	farmerID, err := primitive.ObjectIDFromHex(farmerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID format"})
		return
	}

	file, err := c.FormFile("profilePic")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get profile picture from request: " + err.Error()})
		return
	}

	// Upload to storage service
	fileURL, err := fc.storageService.SaveFile(file, "profiles")
	if err != nil {
		log.Printf("ERROR: Failed to save profile pic to storage for farmer %s: %v", farmerIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload profile picture"})
		return
	}

	// Update DB
	if err := fc.farmerRepo.UpdateProfilePic(farmerID, fileURL); err != nil {
		log.Printf("ERROR: Failed to update profile pic in DB for farmer %s: %v", farmerIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile picture in database"})
		return
	}

	log.Printf("INFO: Profile picture updated — farmer=%s url=%s", farmerIDStr, fileURL)
	c.JSON(http.StatusOK, gin.H{
		"message":    "Profile picture updated successfully",
		"profilePic": fileURL,
	})
}
