// All rights reserved Samyak-Setu

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
		Icon        string `json:"icon"`
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

// GetWeatherDetailed fetches current weather as structured data for API responses.
func (s *OpenWeatherService) GetWeatherDetailed(latitude, longitude float64) (*WeatherData, error) {
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?lat=%.6f&lon=%.6f&appid=%s&units=metric",
		latitude, longitude, s.apiKey,
	)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("weather API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("weather API returned status %d: %s", resp.StatusCode, string(body))
	}

	var weather weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	data := &WeatherData{
		Location:    weather.Name,
		Temperature: weather.Main.Temp,
		FeelsLike:   weather.Main.FeelsLike,
		TempMin:     weather.Main.TempMin,
		TempMax:     weather.Main.TempMax,
		Humidity:    weather.Main.Humidity,
		WindSpeed:   weather.Wind.Speed,
	}

	if len(weather.Weather) > 0 {
		data.Condition = weather.Weather[0].Main
		data.Description = weather.Weather[0].Description
		data.Icon = fmt.Sprintf("https://openweathermap.org/img/wn/%s@2x.png", weather.Weather[0].Icon)
	}

	return data, nil
}

// forecastAPIResponse represents the OpenWeatherMap 5-day/3-hour forecast response.
type forecastAPIResponse struct {
	List []struct {
		DtTxt   string `json:"dt_txt"`
		Weather []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Main struct {
			Temp     float64 `json:"temp"`
			TempMin  float64 `json:"temp_min"`
			TempMax  float64 `json:"temp_max"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	} `json:"list"`
}

// GetForecast fetches a 5-day forecast with 3-hour intervals from OpenWeatherMap.
func (s *OpenWeatherService) GetForecast(latitude, longitude float64) ([]ForecastItem, error) {
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/forecast?lat=%.6f&lon=%.6f&appid=%s&units=metric",
		latitude, longitude, s.apiKey,
	)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("forecast API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("forecast API returned status %d: %s", resp.StatusCode, string(body))
	}

	var forecast forecastAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&forecast); err != nil {
		return nil, fmt.Errorf("failed to decode forecast response: %w", err)
	}

	items := make([]ForecastItem, 0, len(forecast.List))
	for _, entry := range forecast.List {
		item := ForecastItem{
			DateTime:    entry.DtTxt,
			Temperature: entry.Main.Temp,
			TempMin:     entry.Main.TempMin,
			TempMax:     entry.Main.TempMax,
			Humidity:    entry.Main.Humidity,
			WindSpeed:   entry.Wind.Speed,
		}
		if len(entry.Weather) > 0 {
			item.Condition = entry.Weather[0].Main
			item.Description = entry.Weather[0].Description
			item.Icon = fmt.Sprintf("https://openweathermap.org/img/wn/%s@2x.png", entry.Weather[0].Icon)
		}
		items = append(items, item)
	}

	return items, nil
}
