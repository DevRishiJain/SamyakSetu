// All rights reserved Samyak-Setu

package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiService implements AIService using Google's Gemini API.
type GeminiService struct {
	client *genai.Client
}

// NewGeminiService creates a new GeminiService with the provided API key.
func NewGeminiService(apiKey string) (*GeminiService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	log.Println("INFO: Gemini AI client initialized")
	return &GeminiService{client: client}, nil
}

// AnalyzeSoilImage sends a soil image to Gemini Vision and returns the soil type.
func (s *GeminiService) AnalyzeSoilImage(imageData []byte, mimeType string) (string, error) {
	prompt := `You are an expert agricultural soil scientist. Analyze this soil image and identify the soil type.
Respond with ONLY the soil type name (e.g., "Clay", "Sandy", "Loamy", "Silt", "Peat", "Chalky", "Red Soil", "Black Soil", "Alluvial Soil", "Laterite Soil").
If you cannot determine the soil type, respond with "Unknown".
Do not include any other text or explanation.`

	return s.callVisionWithRetry(prompt, imageData, mimeType, 2)
}

// GenerateAdvisory calls Gemini text model for agricultural advice.
func (s *GeminiService) GenerateAdvisory(prompt string) (string, error) {
	return s.callTextWithRetry(prompt, 2)
}

// GenerateAdvisoryWithImage calls Gemini Vision for advice using text + image context.
func (s *GeminiService) GenerateAdvisoryWithImage(prompt string, imageData []byte, mimeType string) (string, error) {
	return s.callVisionWithRetry(prompt, imageData, mimeType, 2)
}

// Close releases the Gemini client resources.
func (s *GeminiService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

// callTextWithRetry calls the Gemini text model with retry logic.
func (s *GeminiService) callTextWithRetry(prompt string, maxRetries int) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("INFO: Gemini text retry attempt %d/%d", attempt, maxRetries)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		model := s.client.GenerativeModel("gemini-2.0-flash")
		model.SetTemperature(0.7)
		model.SetTopP(0.9)

		resp, err := model.GenerateContent(ctx, genai.Text(prompt))
		cancel()

		if err != nil {
			lastErr = fmt.Errorf("gemini text API error: %w", err)
			continue
		}

		text := extractText(resp)
		if text != "" {
			return text, nil
		}

		lastErr = fmt.Errorf("gemini returned empty response")
	}

	return "", fmt.Errorf("gemini text failed after %d retries: %w", maxRetries, lastErr)
}

// callVisionWithRetry calls the Gemini vision model with retry logic.
func (s *GeminiService) callVisionWithRetry(prompt string, imageData []byte, mimeType string, maxRetries int) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("INFO: Gemini vision retry attempt %d/%d", attempt, maxRetries)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		model := s.client.GenerativeModel("gemini-2.0-flash")
		model.SetTemperature(0.4)

		imgPart := genai.ImageData(mimeType, imageData)
		resp, err := model.GenerateContent(ctx, imgPart, genai.Text(prompt))
		cancel()

		if err != nil {
			lastErr = fmt.Errorf("gemini vision API error: %w", err)
			continue
		}

		text := extractText(resp)
		if text != "" {
			return text, nil
		}

		lastErr = fmt.Errorf("gemini vision returned empty response")
	}

	return "", fmt.Errorf("gemini vision failed after %d retries: %w", maxRetries, lastErr)
}

// extractText pulls text content from a Gemini API response.
func extractText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}

	var parts []string
	for _, candidate := range resp.Candidates {
		if candidate.Content == nil {
			continue
		}
		for _, part := range candidate.Content.Parts {
			if textPart, ok := part.(genai.Text); ok {
				parts = append(parts, string(textPart))
			}
		}
	}

	return strings.TrimSpace(strings.Join(parts, "\n"))
}
