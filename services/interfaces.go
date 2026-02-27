package services

import "mime/multipart"

// AIService defines the contract for any AI provider (Gemini, Bedrock, etc.).
type AIService interface {
	// AnalyzeSoilImage sends an image to the AI and returns the identified soil type.
	AnalyzeSoilImage(imageData []byte, mimeType string) (string, error)

	// GenerateAdvisory creates an agricultural advisory response based on context.
	GenerateAdvisory(prompt string) (string, error)

	// GenerateAdvisoryWithImage creates an advisory response using both text and image.
	GenerateAdvisoryWithImage(prompt string, imageData []byte, mimeType string) (string, error)
}

// WeatherService defines the contract for any weather data provider.
type WeatherService interface {
	// GetWeather fetches current weather data for given coordinates.
	// Returns a human-readable weather summary string.
	GetWeather(latitude, longitude float64) (string, error)
}

// StorageService defines the contract for any file storage provider (local, S3, etc.).
type StorageService interface {
	// SaveFile stores an uploaded file and returns the stored file path.
	SaveFile(file *multipart.FileHeader, subDir string) (string, error)
}
