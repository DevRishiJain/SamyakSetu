// All rights reserved Samyak-Setu

package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/services"
)

// VoiceController handles the Voice endpoints for Text-to-Speech (Polly).
type VoiceController struct {
	voiceService services.VoiceService
}

// NewVoiceController creates a new VoiceController instance.
func NewVoiceController(voiceService services.VoiceService) *VoiceController {
	return &VoiceController{
		voiceService: voiceService,
	}
}

// ttsRequest is the expected JSON body for the TTS endpoint.
type ttsRequest struct {
	Text string `json:"text" binding:"required"`
}

// TextToSpeech handles POST /api/voice/tts
// It takes raw Hindi, English or Hinglish text and returns the public Amazon S3 audio link.
func (vc *VoiceController) TextToSpeech(c *gin.Context) {
	var req ttsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if len(req.Text) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text cannot be empty"})
		return
	}

	audioURL, err := vc.voiceService.TextToSpeech(req.Text)
	if err != nil {
		log.Printf("ERROR: TTS failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Text-to-Speech service is temporarily unavailable."})
		return
	}

	log.Printf("INFO: TTS audio generated — len=%d url=%s", len(req.Text), audioURL)
	c.JSON(http.StatusOK, gin.H{"audioUrl": audioURL})
}
