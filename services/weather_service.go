package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// OpenWeatherService implements WeatherService using the OpenWeatherMap API.
type OpenWeatherService struct {
	apiKey     string
	httpClient *http.Client
}

// weatherAPIResponse represents the relevant fields from OpenWeatherMap's response.
type weatherAPIResponse struct {
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Name string `json:"name"`
}

// NewOpenWeatherService creates a new OpenWeatherService instance.
func NewOpenWeatherService(apiKey string) *OpenWeatherService {
	return &OpenWeatherService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetWeather fetches current weather for the given coordinates and returns a summary.
func (s *OpenWeatherService) GetWeather(latitude, longitude float64) (string, error) {
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?lat=%.6f&lon=%.6f&appid=%s&units=metric",
		latitude, longitude, s.apiKey,
	)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		log.Printf("ERROR: Weather API request failed: %v", err)
		return "Weather data unavailable", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("ERROR: Weather API returned status %d: %s", resp.StatusCode, string(body))
		return "Weather data unavailable", nil
	}

	var weather weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		log.Printf("ERROR: Failed to decode weather response: %v", err)
		return "Weather data unavailable", nil
	}

	// Build human-readable summary
	description := "unknown"
	if len(weather.Weather) > 0 {
		description = weather.Weather[0].Description
	}

	summary := fmt.Sprintf(
		"Location: %s | Condition: %s | Temperature: %.1f°C (feels like %.1f°C) | Min: %.1f°C, Max: %.1f°C | Humidity: %d%% | Wind: %.1f m/s",
		weather.Name,
		description,
		weather.Main.Temp,
		weather.Main.FeelsLike,
		weather.Main.TempMin,
		weather.Main.TempMax,
		weather.Main.Humidity,
		weather.Wind.Speed,
	)

	return summary, nil
}
