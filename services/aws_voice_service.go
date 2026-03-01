// All rights reserved Samyak-Setu

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	pollyTypes "github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	transcribeTypes "github.com/aws/aws-sdk-go-v2/service/transcribe/types"
)

// AWSVoiceService implements VoiceService using Amazon Polly (TTS) and Amazon Transcribe (STT).
type AWSVoiceService struct {
	pollyClient      *polly.Client
	transcribeClient *transcribe.Client
	storageService   StorageService
	s3BucketName     string
	s3Region         string
}

// NewAWSVoiceService initializes a new AWSVoiceService with Polly and Transcribe.
func NewAWSVoiceService(region, accessKey, secretKey, s3BucketName string, storageService StorageService) (*AWSVoiceService, error) {
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")

	// Polly uses ap-south-1 (Mumbai) because Kajal Neural voice is widely supported there,
	// while it throws "engine not supported" in eu-north-1 (Stockholm).
	pollyCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ap-south-1"),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for Polly: %w", err)
	}

	// Transcribe uses the same region as S3 so it can read the uploaded audio file
	// from the same bucket without cross-region access issues.
	transcribeCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for Transcribe: %w", err)
	}

	pollyClient := polly.NewFromConfig(pollyCfg)
	transcribeClient := transcribe.NewFromConfig(transcribeCfg)
	log.Printf("INFO: AWS Voice Service initialized (Polly in ap-south-1, Transcribe in %s)", region)

	return &AWSVoiceService{
		pollyClient:      pollyClient,
		transcribeClient: transcribeClient,
		storageService:   storageService,
		s3BucketName:     s3BucketName,
		s3Region:         region,
	}, nil
}

// TextToSpeech converts text into an MP3 file using Amazon Polly, uploads to S3, and returns the public URL.
func (s *AWSVoiceService) TextToSpeech(text string) (string, error) {
	// Kajal is a high-quality Indian Neural voice — sounds like a real person, not robotic.
	// She can speak Hindi, English, and Hinglish seamlessly.
	input := &polly.SynthesizeSpeechInput{
		OutputFormat: pollyTypes.OutputFormatMp3,
		Text:         aws.String(text),
		VoiceId:      pollyTypes.VoiceIdKajal,
		Engine:       pollyTypes.EngineNeural,
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

	publicURL, err := s.storageService.SaveBytes(audioBytes, "audio/mpeg", ".mp3", "audio")
	if err != nil {
		return "", fmt.Errorf("failed to save audio to storage: %w", err)
	}

	return publicURL, nil
}

// SpeechToText converts an audio file to text using Amazon Transcribe.
// It uploads the audio to S3, starts a transcription job, polls for completion, and returns the transcribed text.
func (s *AWSVoiceService) SpeechToText(audioData []byte, ext string) (string, error) {
	// 1. Determine content type and media format
	contentType := "audio/wav"
	mediaFormat := transcribeTypes.MediaFormatWav
	if strings.HasSuffix(ext, "mp3") {
		contentType = "audio/mpeg"
		mediaFormat = transcribeTypes.MediaFormatMp3
	} else if strings.HasSuffix(ext, "mp4") || strings.HasSuffix(ext, "m4a") {
		contentType = "audio/mp4"
		mediaFormat = transcribeTypes.MediaFormatMp4
	} else if strings.HasSuffix(ext, "ogg") {
		contentType = "audio/ogg"
		mediaFormat = transcribeTypes.MediaFormatOgg
	} else if strings.HasSuffix(ext, "flac") {
		contentType = "audio/flac"
		mediaFormat = transcribeTypes.MediaFormatFlac
	} else if strings.HasSuffix(ext, "webm") {
		contentType = "audio/webm"
		mediaFormat = transcribeTypes.MediaFormatWebm
	}

	// 2. Upload audio to S3 so Transcribe can access it
	s3URL, err := s.storageService.SaveBytes(audioData, contentType, ext, "stt-input")
	if err != nil {
		return "", fmt.Errorf("failed to upload audio to S3: %w", err)
	}

	// 3. Start transcription job with a unique name
	jobName := fmt.Sprintf("samyak-stt-%d", time.Now().UnixNano())
	_, err = s.transcribeClient.StartTranscriptionJob(context.TODO(), &transcribe.StartTranscriptionJobInput{
		TranscriptionJobName: aws.String(jobName),
		Media: &transcribeTypes.Media{
			MediaFileUri: aws.String(s3URL),
		},
		MediaFormat:      mediaFormat,
		IdentifyLanguage: aws.Bool(true), // Auto-detect Hindi / English
		LanguageOptions:  []transcribeTypes.LanguageCode{transcribeTypes.LanguageCodeHiIn, transcribeTypes.LanguageCodeEnUs, transcribeTypes.LanguageCodeEnIn},
	})
	if err != nil {
		return "", fmt.Errorf("failed to start transcription job: %w", err)
	}

	log.Printf("INFO: Transcription job started — name=%s", jobName)

	// 4. Poll until the job completes (max ~90 seconds)
	var transcriptURI string
	for i := 0; i < 30; i++ {
		time.Sleep(3 * time.Second)

		result, err := s.transcribeClient.GetTranscriptionJob(context.TODO(), &transcribe.GetTranscriptionJobInput{
			TranscriptionJobName: aws.String(jobName),
		})
		if err != nil {
			return "", fmt.Errorf("failed to get transcription job status: %w", err)
		}

		status := result.TranscriptionJob.TranscriptionJobStatus
		if status == transcribeTypes.TranscriptionJobStatusCompleted {
			transcriptURI = *result.TranscriptionJob.Transcript.TranscriptFileUri
			log.Printf("INFO: Transcription job completed — name=%s", jobName)
			break
		} else if status == transcribeTypes.TranscriptionJobStatusFailed {
			reason := ""
			if result.TranscriptionJob.FailureReason != nil {
				reason = *result.TranscriptionJob.FailureReason
			}
			return "", fmt.Errorf("transcription job failed: %s", reason)
		}
		// Still IN_PROGRESS — keep polling
	}

	if transcriptURI == "" {
		return "", fmt.Errorf("transcription job timed out after 90 seconds")
	}

	// 5. Download the JSON transcript
	resp, err := http.Get(transcriptURI)
	if err != nil {
		return "", fmt.Errorf("failed to download transcript: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read transcript body: %w", err)
	}

	// 6. Parse the transcript JSON and extract the text
	var transcriptResult struct {
		Results struct {
			Transcripts []struct {
				Transcript string `json:"transcript"`
			} `json:"transcripts"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &transcriptResult); err != nil {
		return "", fmt.Errorf("failed to parse transcript JSON: %w", err)
	}

	if len(transcriptResult.Results.Transcripts) == 0 {
		return "", fmt.Errorf("transcription returned no text")
	}

	text := transcriptResult.Results.Transcripts[0].Transcript
	if text == "" {
		return "", fmt.Errorf("transcription returned empty text")
	}

	log.Printf("INFO: Transcription complete — text_len=%d text=%s", len(text), text)

	// 7. Clean up the transcription job (best effort, don't fail if this errors)
	_, _ = s.transcribeClient.DeleteTranscriptionJob(context.TODO(), &transcribe.DeleteTranscriptionJobInput{
		TranscriptionJobName: aws.String(jobName),
	})

	return text, nil
}
