// All rights reserved Samyak-Setu

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

// WeatherData holds structured weather information for API responses.
type WeatherData struct {
	Location    string  `json:"location"`
	Condition   string  `json:"condition"`
	Description string  `json:"description"`
	Temperature float64 `json:"temperature"`
	FeelsLike   float64 `json:"feelsLike"`
	TempMin     float64 `json:"tempMin"`
	TempMax     float64 `json:"tempMax"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"windSpeed"`
	Icon        string  `json:"icon"`
}

// ForecastItem holds weather data for a single forecast time slot.
type ForecastItem struct {
	DateTime    string  `json:"dateTime"`
	Condition   string  `json:"condition"`
	Description string  `json:"description"`
	Temperature float64 `json:"temperature"`
	TempMin     float64 `json:"tempMin"`
	TempMax     float64 `json:"tempMax"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"windSpeed"`
	Icon        string  `json:"icon"`
}

// WeatherService defines the contract for any weather data provider.
type WeatherService interface {
	// GetWeather fetches current weather data for given coordinates.
	// Returns a human-readable weather summary string.
	GetWeather(latitude, longitude float64) (string, error)

	// GetWeatherDetailed fetches current weather as structured data.
	GetWeatherDetailed(latitude, longitude float64) (*WeatherData, error)

	// GetForecast fetches a 5-day / 3-hour forecast for the given coordinates.
	GetForecast(latitude, longitude float64) ([]ForecastItem, error)
}

// StorageService defines the contract for any file storage provider (local, S3, etc.).
type StorageService interface {
	// SaveFile stores an uploaded file and returns the stored file path.
	SaveFile(file *multipart.FileHeader, subDir string) (string, error)
}
