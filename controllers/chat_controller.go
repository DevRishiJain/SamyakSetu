// All rights reserved Samyak-Setu

package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/models"
	"github.com/samyaksetu/backend/repositories"
	"github.com/samyaksetu/backend/services"
	"github.com/samyaksetu/backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ChatController handles HTTP requests related to AI advisory chat.
type ChatController struct {
	farmerRepo     *repositories.FarmerRepository
	soilRepo       *repositories.SoilRepository
	chatRepo       *repositories.ChatRepository
	aiService      services.AIService
	weatherService services.WeatherService
}

// NewChatController creates a new ChatController instance.
func NewChatController(
	farmerRepo *repositories.FarmerRepository,
	soilRepo *repositories.SoilRepository,
	chatRepo *repositories.ChatRepository,
	aiService services.AIService,
	weatherService services.WeatherService,
) *ChatController {
	return &ChatController{
		farmerRepo:     farmerRepo,
		soilRepo:       soilRepo,
		chatRepo:       chatRepo,
		aiService:      aiService,
		weatherService: weatherService,
	}
}

// Chat handles POST /api/chat — AI-powered agricultural advisory.
func (cc *ChatController) Chat(c *gin.Context) {
	// Parse farmer ID
	farmerIDStr := c.PostForm("farmerId")
	if farmerIDStr == "" {
		// Try JSON binding as fallback
		farmerIDStr = c.Query("farmerId")
	}

	// Also try from JSON body
	var jsonReq models.ChatRequest
	if farmerIDStr == "" {
		if err := c.ShouldBindJSON(&jsonReq); err == nil {
			farmerIDStr = jsonReq.FarmerID
		}
	} else {
		jsonReq.Message = c.PostForm("message")
	}

	if farmerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "farmerId is required"})
		return
	}

	farmerID, err := primitive.ObjectIDFromHex(farmerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID format"})
		return
	}

	// Get message
	message := jsonReq.Message
	if message == "" {
		message = c.PostForm("message")
	}
	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	// Ensure farmer exists
	farmer, err := cc.farmerRepo.FindByID(farmerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Farmer not found"})
		return
	}

	// Fetch latest soil data (optional — farmer may not have uploaded soil yet)
	soilType := "Not available (no soil analysis done yet)"
	soilData, err := cc.soilRepo.FindLatestByFarmerID(farmerID)
	if err == nil && soilData != nil {
		soilType = soilData.SoilType
	} else if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("WARN: Failed to fetch soil data for farmer %s: %v", farmerID.Hex(), err)
	}

	// Fetch weather data
	weatherSummary, err := cc.weatherService.GetWeather(farmer.Location.Latitude, farmer.Location.Longitude)
	if err != nil {
		log.Printf("WARN: Weather fetch failed for farmer %s: %v", farmerID.Hex(), err)
		weatherSummary = "Weather data unavailable"
	}

	// Check for optional image
	var imageData []byte
	var mimeType string
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		if validErr := utils.ValidateImageFile(file); validErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validErr.Error()})
			return
		}
		src, openErr := file.Open()
		if openErr == nil {
			imageData, _ = io.ReadAll(src)
			src.Close()
			mimeType = utils.GetMimeType(file)
		}
	}

	// Build structured prompt
	prompt := buildAdvisoryPrompt(farmer, soilType, weatherSummary, message)

	// Save user message
	userMsg := &models.ChatMessage{
		FarmerID: farmerID,
		Role:     "user",
		Message:  message,
	}
	if file != nil {
		userMsg.ImagePath = file.Filename
	}
	if saveErr := cc.chatRepo.SaveMessage(userMsg); saveErr != nil {
		log.Printf("WARN: Failed to save user message: %v", saveErr)
	}

	// Call AI
	var aiReply string
	if len(imageData) > 0 {
		aiReply, err = cc.aiService.GenerateAdvisoryWithImage(prompt, imageData, mimeType)
	} else {
		aiReply, err = cc.aiService.GenerateAdvisory(prompt)
	}

	if err != nil {
		log.Printf("ERROR: AI advisory failed for farmer %s: %v", farmerID.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI advisory service is temporarily unavailable. Please try again."})
		return
	}

	// Save AI response
	aiMsg := &models.ChatMessage{
		FarmerID: farmerID,
		Role:     "ai",
		Message:  aiReply,
	}
	if saveErr := cc.chatRepo.SaveMessage(aiMsg); saveErr != nil {
		log.Printf("WARN: Failed to save AI message: %v", saveErr)
	}

	log.Printf("INFO: Chat completed — farmer=%s query_len=%d reply_len=%d", farmer.Name, len(message), len(aiReply))
	c.JSON(http.StatusOK, models.ChatResponse{Reply: aiReply})
}

// buildAdvisoryPrompt constructs a context-rich prompt for agricultural advisory.
func buildAdvisoryPrompt(farmer *models.Farmer, soilType, weather, query string) string {
	return fmt.Sprintf(`You are SamyakSetu AI, an expert agricultural advisor for Indian farmers.
You provide practical, actionable advice based on the farmer's specific conditions.

=== FARMER CONTEXT ===
Name: %s
Location: Latitude %.6f, Longitude %.6f
Soil Type: %s
Current Weather: %s

=== FARMER'S QUESTION ===
%s

=== INSTRUCTIONS ===
1. Provide advice specific to the farmer's soil type, location, and current weather conditions.
2. If the farmer asks about crops, recommend varieties suitable for their soil and climate.
3. If asking about pests or diseases, consider the weather conditions in your diagnosis.
4. Keep advice practical and actionable for a small to medium-scale farmer.
5. If relevant, mention any weather-related precautions.
6. Respond in a friendly, supportive tone.
7. If you don't have enough context, ask clarifying questions.
8. Keep the response concise but comprehensive (200-400 words unless more detail is needed).`,
		farmer.Name,
		farmer.Location.Latitude,
		farmer.Location.Longitude,
		soilType,
		weather,
		query,
	)
}
