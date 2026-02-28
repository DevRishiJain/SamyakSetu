// All rights reserved Samyak-Setu

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values loaded from environment variables.
type Config struct {
	Port                string
	MongoURI            string
	GeminiAPIKey        string
	WeatherAPIKey       string
	UploadPath          string
	AWSRegion           string
	AWSAccessKey        string
	AWSSecretKey        string
	S3BucketName        string
	BedrockRegion       string
	BedrockAccessKey    string
	BedrockSecretKey    string
	BedrockSessionToken string // In case you use temporary credentials, usually empty for IAM users
}

// LoadConfig reads the .env file and returns a Config struct.
// Falls back to OS environment variables if .env is not found.
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("WARN: .env file not found, using OS environment variables")
	}

	cfg := &Config{
		Port:                getEnv("PORT", "8080"),
		MongoURI:            getEnv("MONGO_URI", "mongodb://localhost:27017/samyaksetu"),
		GeminiAPIKey:        getEnv("GEMINI_API_KEY", ""),
		WeatherAPIKey:       getEnv("WEATHER_API_KEY", ""),
		UploadPath:          getEnv("UPLOAD_PATH", "./uploads"),
		AWSRegion:           getEnv("AWS_REGION", ""),
		AWSAccessKey:        getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretKey:        getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3BucketName:        getEnv("S3_BUCKET_NAME", ""),
		BedrockRegion:       getEnv("BEDROCK_AWS_REGION", "us-east-1"), // Defaulting to us-east-1 since many models are there
		BedrockAccessKey:    getEnv("BEDROCK_AWS_ACCESS_KEY_ID", ""),
		BedrockSecretKey:    getEnv("BEDROCK_AWS_SECRET_ACCESS_KEY", ""),
		BedrockSessionToken: getEnv("BEDROCK_AWS_SESSION_TOKEN", ""), // Optional
	}

	if cfg.GeminiAPIKey == "" && (cfg.BedrockAccessKey == "" || cfg.BedrockSecretKey == "") {
		log.Println("WARN: Neither GEMINI_API_KEY nor AWS Bedrock credentials are set — AI features will fail")
	}
	if cfg.WeatherAPIKey == "" {
		log.Println("WARN: WEATHER_API_KEY is not set — weather features will fail")
	}

	return cfg
}

// getEnv returns the value of an environment variable or a fallback default.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
