// All rights reserved Samyak-Setu

package services

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
)

// AWSVoiceService implements VoiceService using Amazon Polly.
type AWSVoiceService struct {
	pollyClient    *polly.Client
	storageService StorageService
}

// NewAWSVoiceService initializes a new AWSVoiceService with Polly.
func NewAWSVoiceService(region, accessKey, secretKey string, storageService StorageService) (*AWSVoiceService, error) {
	// Force region to ap-south-1 (Mumbai) because Kajal Neural voice is widely supported there,
	// while it might throw "engine not supported" in regions like eu-north-1 (Stockholm).
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ap-south-1"),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for Voice Service: %w", err)
	}

	pollyClient := polly.NewFromConfig(cfg)
	log.Println("INFO: AWS Voice Service initialized (Amazon Polly) in ap-south-1")

	return &AWSVoiceService{
		pollyClient:    pollyClient,
		storageService: storageService,
	}, nil
}

// TextToSpeech converts text into an MP3 file using Amazon Polly, uploads to storage, and returns the URL.
func (s *AWSVoiceService) TextToSpeech(text string) (string, error) {
	// We use Kajal, the very natural sounding Indian Neural Voice that speaks both Hindi and Indian English seamlessly.
	input := &polly.SynthesizeSpeechInput{
		OutputFormat: types.OutputFormatMp3,
		Text:         aws.String(text),
		VoiceId:      types.VoiceIdKajal,
		Engine:       types.EngineNeural,
	}

	out, err := s.pollyClient.SynthesizeSpeech(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("failed to synthesize speech: %w", err)
	}
	defer out.AudioStream.Close()

	audioBytes, err := io.ReadAll(out.AudioStream)
	if err != nil {
		return "", fmt.Errorf("failed to read audio stream: %w", err)
	}

	// Use StorageService to instantly save this to S3 and get the playback URL
	publicURL, err := s.storageService.SaveBytes(audioBytes, "audio/mpeg", ".mp3", "audio")
	if err != nil {
		return "", fmt.Errorf("failed to save audio stream to storage: %w", err)
	}

	return publicURL, nil
}

// SpeechToText is deliberately left to the frontend for zero-latency.
func (s *AWSVoiceService) SpeechToText(audioData []byte, ext string) (string, error) {
	// The frontend (React / Flutter / Native) should just trigger the user's native device microphone (Gboard / Siri)
	return "", fmt.Errorf("unsupported: use native frontend speech-to-text keyboards for strict zero-latency typing")
}
