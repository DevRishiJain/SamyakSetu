// All rights reserved Samyak-Setu

package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/services"
)

// SamyakAIController handles the standalone AI chatbot for farmers.
type SamyakAIController struct {
	aiService services.AIService
}

// NewSamyakAIController creates a new SamyakAIController instance.
func NewSamyakAIController(aiService services.AIService) *SamyakAIController {
	return &SamyakAIController{
		aiService: aiService,
	}
}

// samyakAIRequest is the expected JSON body for the /api/samyakai endpoint.
type samyakAIRequest struct {
	Message string `json:"message" binding:"required"`
}

// Chat handles POST /api/samyakai — an open conversational AI for farmers.
// It only answers farming, crops, agriculture, soil, weather, irrigation,
// livestock, farmer mental health, and related topics.
func (sc *SamyakAIController) Chat(c *gin.Context) {
	var req samyakAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if len(req.Message) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message cannot be empty"})
		return
	}

	prompt := buildSamyakAIPrompt(req.Message)

	reply, err := sc.aiService.GenerateAdvisory(prompt)
	if err != nil {
		log.Printf("ERROR: SamyakAI chat failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service is temporarily unavailable. Please try again."})
		return
	}

	log.Printf("INFO: SamyakAI chat — query_len=%d reply_len=%d", len(req.Message), len(reply))
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

// buildSamyakAIPrompt constructs the system prompt that restricts the AI to farming topics only.
func buildSamyakAIPrompt(userMessage string) string {
	return fmt.Sprintf(`You are SamyakAI, a friendly and expert agricultural chatbot built for Indian farmers.
You are like a wise elder farmer who has decades of experience and speaks warmly.

=== STRICT RULES ===
1. You MUST ONLY answer questions related to:
   - Farming, agriculture, crops, seeds, harvesting, irrigation, fertilizers
   - Soil health, soil types, soil testing, composting
   - Weather impact on farming, seasonal crop planning
   - Pest control, plant diseases, weed management
   - Livestock, dairy farming, poultry, fisheries
   - Government farming schemes, MSP (Minimum Support Price), crop insurance
   - Farmer mental health, stress, loneliness, financial anxiety, emotional support
   - Farm equipment, modern farming techniques, organic farming
   - Market prices, mandi rates, selling strategies
   - Water management, drip irrigation, rainwater harvesting

2. If a user asks about ANY topic NOT related to farming or farmer welfare, you MUST politely refuse.
   Example refusal: "Namaste! I am SamyakAI, your farming companion. I can only help with agriculture, crops, farming techniques, and farmer well-being. Please ask me something related to farming! 🌾"

3. You may respond in Hindi, English, or Hinglish depending on how the user writes.
4. Keep responses practical, actionable, and easy to understand for a farmer.
5. Be warm, supportive, and encouraging. Farmers work incredibly hard.

=== FARMER'S MESSAGE ===
%s

=== RESPOND NOW ===`, userMessage)
}
