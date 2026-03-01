// All rights reserved Samyak-Setu

package controllers

import (
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/services"
)

// VoiceController handles the Voice endpoints for TTS (Polly) and STT (Transcribe).
type VoiceController struct {
	voiceService services.VoiceService
	aiService    services.AIService
}

// NewVoiceController creates a new VoiceController instance.
func NewVoiceController(voiceService services.VoiceService, aiService services.AIService) *VoiceController {
	return &VoiceController{
		voiceService: voiceService,
		aiService:    aiService,
	}
}

// ttsRequest is the expected JSON body for the TTS endpoint.
type ttsRequest struct {
	Text string `json:"text" binding:"required"`
}

// TextToSpeech handles POST /api/voice/tts
// Converts text into a natural-sounding MP3 using Amazon Polly (Kajal Neural voice).
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

// SpeechToText handles POST /api/voice/stt
// Accepts a multipart audio file upload (mp3, wav, m4a, ogg, flac, webm) and returns the transcribed text.
func (vc *VoiceController) SpeechToText(c *gin.Context) {
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'audio' file in request"})
		return
	}
	defer file.Close()

	audioData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read audio file"})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".wav" // default
	}

	log.Printf("INFO: STT request — filename=%s size=%d ext=%s", header.Filename, len(audioData), ext)

	text, err := vc.voiceService.SpeechToText(audioData, ext)
	if err != nil {
		log.Printf("ERROR: STT failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Speech-to-Text service failed: " + err.Error()})
		return
	}

	log.Printf("INFO: STT completed — text=%s", text)
	c.JSON(http.StatusOK, gin.H{"text": text})
}

// VoiceChat handles POST /api/voice/chat
// Full voice-to-voice pipeline:
//  1. Farmer sends audio file (wav, mp3, etc.)
//  2. Backend transcribes it to text (Amazon Transcribe)
//  3. Backend sends the text to SamyakAI for a farming-focused response
//  4. Backend converts the AI response to speech (Amazon Polly)
//  5. Returns both the text reply AND the audio URL
func (vc *VoiceController) VoiceChat(c *gin.Context) {
	// Step 1: Read audio file
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'audio' file in request"})
		return
	}
	defer file.Close()

	audioData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read audio file"})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".wav"
	}

	log.Printf("INFO: VoiceChat request — filename=%s size=%d", header.Filename, len(audioData))

	// Step 2: Transcribe audio to text
	userText, err := vc.voiceService.SpeechToText(audioData, ext)
	if err != nil {
		log.Printf("ERROR: VoiceChat STT failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not understand the audio. Please try again."})
		return
	}

	log.Printf("INFO: VoiceChat transcribed — text=%s", userText)

	// Step 3: Send to SamyakAI (farming-focused chatbot)
	prompt := buildVoiceChatPrompt(userText)
	aiReply, err := vc.aiService.GenerateAdvisory(prompt)
	if err != nil {
		log.Printf("ERROR: VoiceChat AI failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service is temporarily unavailable."})
		return
	}

	log.Printf("INFO: VoiceChat AI replied — reply_len=%d", len(aiReply))

	// Step 4: Convert AI response to speech
	audioURL, err := vc.voiceService.TextToSpeech(aiReply)
	if err != nil {
		log.Printf("ERROR: VoiceChat TTS failed: %v", err)
		// Still return the text reply even if audio generation fails
		c.JSON(http.StatusOK, gin.H{
			"userText": userText,
			"reply":    aiReply,
			"audioUrl": nil,
			"error":    "Audio generation failed, but text reply is available.",
		})
		return
	}

	log.Printf("INFO: VoiceChat complete — user=%s reply_len=%d audio=%s", userText, len(aiReply), audioURL)

	// Step 5: Return everything
	c.JSON(http.StatusOK, gin.H{
		"userText": userText,
		"reply":    aiReply,
		"audioUrl": audioURL,
	})
}

// buildVoiceChatPrompt constructs the SamyakAI system prompt for voice chat.
func buildVoiceChatPrompt(userMessage string) string {
	return `You are SamyakAI, a friendly and expert agricultural chatbot built for Indian farmers.
You are speaking to the farmer through VOICE, so keep your responses concise and conversational.

=== STRICT RULES ===
1. You MUST ONLY answer questions related to:
   - Farming, agriculture, crops, seeds, harvesting, irrigation, fertilizers
   - Soil health, soil types, composting, organic farming
   - Weather impact on farming, seasonal crop planning
   - Pest control, plant diseases, weed management
   - Livestock, dairy farming, poultry, fisheries
   - Government farming schemes, MSP, crop insurance
   - Farmer mental health, stress, emotional support
   - Farm equipment, modern farming techniques
   - Market prices, mandi rates, selling strategies

2. If a user asks about ANY topic NOT related to farming, politely refuse.

3. Since this is VOICE conversation, keep answers SHORT and CLEAR (2-3 paragraphs max).
   Avoid bullet points and numbered lists — speak naturally like a conversation.

4. Respond in the same language the farmer used (Hindi, English, or Hinglish).

=== FARMER SAID ===
` + userMessage + `

=== RESPOND NOW (keep it concise for voice) ===`
}
