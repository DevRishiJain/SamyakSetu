// All rights reserved Samyak-Setu

package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/repositories"
	"github.com/samyaksetu/backend/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WeatherController handles HTTP requests related to weather data.
type WeatherController struct {
	farmerRepo     *repositories.FarmerRepository
	weatherService services.WeatherService
}

// NewWeatherController creates a new WeatherController instance.
func NewWeatherController(farmerRepo *repositories.FarmerRepository, weatherService services.WeatherService) *WeatherController {
	return &WeatherController{
		farmerRepo:     farmerRepo,
		weatherService: weatherService,
	}
}

// GetWeather handles GET /api/weather?farmerId=xxx — returns current weather + 5-day forecast.
func (wc *WeatherController) GetWeather(c *gin.Context) {
	farmerIDStr := c.Query("farmerId")
	if farmerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "farmerId query parameter is required"})
		return
	}

	farmerID, err := primitive.ObjectIDFromHex(farmerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID format"})
		return
	}

	farmer, err := wc.farmerRepo.FindByID(farmerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Farmer not found"})
		return
	}

	// Fetch current weather (structured)
	current, err := wc.weatherService.GetWeatherDetailed(farmer.Location.Latitude, farmer.Location.Longitude)
	if err != nil {
		log.Printf("ERROR: Weather fetch failed for farmer %s: %v", farmerID.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather data"})
		return
	}

	// Fetch 5-day forecast
	forecast, err := wc.weatherService.GetForecast(farmer.Location.Latitude, farmer.Location.Longitude)
	if err != nil {
		log.Printf("WARN: Forecast fetch failed for farmer %s: %v", farmerID.Hex(), err)
		// Return current weather even if forecast fails
		c.JSON(http.StatusOK, gin.H{
			"farmerId":      farmerID.Hex(),
			"location":      farmer.Location,
			"current":       current,
			"forecast":      []interface{}{},
			"forecastError": "Forecast data temporarily unavailable",
		})
		return
	}

	log.Printf("INFO: Weather data fetched — farmer=%s location=%s", farmer.Name, current.Location)
	c.JSON(http.StatusOK, gin.H{
		"farmerId": farmerID.Hex(),
		"location": farmer.Location,
		"current":  current,
		"forecast": forecast,
	})
}
