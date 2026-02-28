// All rights reserved Samyak-Setu

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// BedrockService implements AIService using AWS Bedrock (Amazon Nova).
type BedrockService struct {
	client *bedrockruntime.Client
	model  string
}

// NewBedrockService creates a new BedrockService with the provided AWS credentials and region.
func NewBedrockService(region, accessKey, secretKey, sessionToken string) (*BedrockService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
				SessionToken:    sessionToken,
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load bedrock config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)
	log.Println("INFO: AWS Bedrock client initialized (Amazon Nova Lite)")

	return &BedrockService{
		client: client,
		model:  "amazon.nova-lite-v1:0",
	}, nil
}

// AnalyzeSoilImage sends a soil image to Amazon Nova and returns the soil type.
func (s *BedrockService) AnalyzeSoilImage(imageData []byte, mimeType string) (string, error) {
	prompt := `You are an expert agricultural soil scientist. Analyze this soil image and identify the soil type.
Respond with ONLY the soil type name (e.g., "Clay", "Sandy", "Loamy", "Silt", "Peat", "Chalky", "Red Soil", "Black Soil", "Alluvial Soil", "Laterite Soil").
If you cannot determine the soil type, respond with "Unknown".
Do not include any other text or explanation.`

	return s.callVisionWithRetry(prompt, imageData, mimeType, 2)
}

// GenerateAdvisory calls Amazon Nova text model for agricultural advice.
func (s *BedrockService) GenerateAdvisory(prompt string) (string, error) {
	return s.callTextWithRetry(prompt, 2)
}

// GenerateAdvisoryWithImage calls Amazon Nova vision model for advice using text + image context.
func (s *BedrockService) GenerateAdvisoryWithImage(prompt string, imageData []byte, mimeType string) (string, error) {
	return s.callVisionWithRetry(prompt, imageData, mimeType, 2)
}

// Close releases any resources if necessary (AWS SDK handles this mostly, but provided to match interface).
func (s *BedrockService) Close() {
	// Not needed for bedrockruntime.Client
}

// --- Bedrock / Amazon Nova specific structs ---

type novaTextContent struct {
	Text string `json:"text"`
}

type novaImageSource struct {
	Bytes []byte `json:"bytes"`
}

type novaImageFormat struct {
	Format string          `json:"format"`
	Source novaImageSource `json:"source"`
}

type novaImageContent struct {
	Image novaImageFormat `json:"image"`
}

type novaMessage struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type novaRequest struct {
	Messages []novaMessage `json:"messages"`
}

type novaResponse struct {
	Output struct {
		Message struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"message"`
	} `json:"output"`
}

// callTextWithRetry calls Amazon Nova text API
func (s *BedrockService) callTextWithRetry(prompt string, maxRetries int) (string, error) {
	reqBody := novaRequest{
		Messages: []novaMessage{
			{
				Role: "user",
				Content: []interface{}{
					novaTextContent{Text: prompt},
				},
			},
		},
	}

	payloadInfo, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("INFO: Bedrock text retry attempt %d/%d", attempt, maxRetries)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		output, err := s.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
			Body:        payloadInfo,
			ModelId:     aws.String(s.model),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
		})
		cancel()

		if err != nil {
			lastErr = fmt.Errorf("bedrock API error: %w", err)
			continue
		}

		var response novaResponse
		if err := json.Unmarshal(output.Body, &response); err != nil {
			lastErr = fmt.Errorf("failed to decode bedrock response: %w", err)
			continue
		}

		if len(response.Output.Message.Content) > 0 {
			return response.Output.Message.Content[0].Text, nil
		}

		lastErr = fmt.Errorf("bedrock returned empty content response")
	}

	return "", fmt.Errorf("bedrock failed after %d retries: %w", maxRetries, lastErr)
}

// callVisionWithRetry calls Amazon Nova vision API
func (s *BedrockService) callVisionWithRetry(prompt string, imageData []byte, mimeType string, maxRetries int) (string, error) {
	format := "png" // default
	if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
		format = "jpeg"
	} else if strings.Contains(mimeType, "webp") {
		format = "webp"
	} else if strings.Contains(mimeType, "gif") {
		format = "gif"
	}

	reqBody := novaRequest{
		Messages: []novaMessage{
			{
				Role: "user",
				Content: []interface{}{
					novaImageContent{
						Image: novaImageFormat{
							Format: format,
							Source: novaImageSource{
								Bytes: imageData,
							},
						},
					},
					novaTextContent{Text: prompt},
				},
			},
		},
	}

	payloadInfo, _ := json.Marshal(reqBody)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("INFO: Bedrock vision retry attempt %d/%d", attempt, maxRetries)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		output, err := s.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
			Body:        payloadInfo,
			ModelId:     aws.String(s.model),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
		})
		cancel()

		if err != nil {
			lastErr = fmt.Errorf("bedrock vision API error: %w", err)
			log.Printf("ERROR: bedrock vision error payload %v", err)
			continue
		}

		var response novaResponse
		if err := json.Unmarshal(output.Body, &response); err != nil {
			lastErr = fmt.Errorf("failed to decode bedrock response: %w", err)
			continue
		}

		if len(response.Output.Message.Content) > 0 {
			return response.Output.Message.Content[0].Text, nil
		}

		lastErr = fmt.Errorf("bedrock returned empty content response")
	}

	return "", fmt.Errorf("bedrock vision failed after %d retries: %w", maxRetries, lastErr)
}
