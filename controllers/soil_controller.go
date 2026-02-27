// All rights reserved Samyak-Setu

package controllers

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/models"
	"github.com/samyaksetu/backend/repositories"
	"github.com/samyaksetu/backend/services"
	"github.com/samyaksetu/backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SoilController handles HTTP requests related to soil analysis.
type SoilController struct {
	farmerRepo     *repositories.FarmerRepository
	soilRepo       *repositories.SoilRepository
	aiService      services.AIService
	storageService services.StorageService
}

// NewSoilController creates a new SoilController instance.
func NewSoilController(
	farmerRepo *repositories.FarmerRepository,
	soilRepo *repositories.SoilRepository,
	aiService services.AIService,
	storageService services.StorageService,
) *SoilController {
	return &SoilController{
		farmerRepo:     farmerRepo,
		soilRepo:       soilRepo,
		aiService:      aiService,
		storageService: storageService,
	}
}

// UploadSoil handles POST /api/soil/upload — uploads a soil image and analyzes it.
func (sc *SoilController) UploadSoil(c *gin.Context) {
	// Get farmer ID from form
	farmerIDStr := c.PostForm("farmerId")
	if farmerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "farmerId is required"})
		return
	}

	farmerID, err := primitive.ObjectIDFromHex(farmerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID format"})
		return
	}

	// Ensure farmer exists
	_, err = sc.farmerRepo.FindByID(farmerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Farmer not found"})
		return
	}

	// Get uploaded file
	file, err := c.FormFile("soilImage")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "soilImage is required"})
		return
	}

	// Validate image file
	if err := utils.ValidateImageFile(file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Read image bytes for AI analysis
	src, err := file.Open()
	if err != nil {
		log.Printf("ERROR: Failed to open uploaded file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image"})
		return
	}
	imageData, err := io.ReadAll(src)
	src.Close()
	if err != nil {
		log.Printf("ERROR: Failed to read uploaded file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
		return
	}

	// Save file to storage
	storedPath, err := sc.storageService.SaveFile(file, "soil")
	if err != nil {
		log.Printf("ERROR: Failed to save soil image: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Analyze soil with AI
	mimeType := utils.GetMimeType(file)
	soilType, err := sc.aiService.AnalyzeSoilImage(imageData, mimeType)
	if err != nil {
		log.Printf("WARN: AI soil analysis failed: %v", err)
		soilType = "Unknown (Pending AI Analysis)"
	}

	// Save soil data to database
	soilData := &models.SoilData{
		FarmerID:  farmerID,
		ImagePath: storedPath,
		SoilType:  soilType,
	}

	if err := sc.soilRepo.Create(soilData); err != nil {
		log.Printf("ERROR: Failed to save soil data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save soil data"})
		return
	}

	log.Printf("INFO: Soil analyzed — farmer=%s soilType=%s path=%s", farmerID.Hex(), soilType, storedPath)
	c.JSON(http.StatusOK, models.SoilUploadResponse{
		SoilType:  soilType,
		ImagePath: storedPath,
	})
}
