// All rights reserved Samyak-Setu

package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/controllers"
	"github.com/samyaksetu/backend/middlewares"
	"github.com/samyaksetu/backend/services"
)

// RegisterRoutes sets up all API routes on the Gin engine.
func RegisterRoutes(
	router *gin.Engine,
	authCtrl *controllers.AuthController,
	farmerCtrl *controllers.FarmerController,
	soilCtrl *controllers.SoilController,
	chatCtrl *controllers.ChatController,
	weatherCtrl *controllers.WeatherController,
	samyakAICtrl *controllers.SamyakAIController,
	voiceCtrl *controllers.VoiceController,
	jwtService *services.JWTService,
) {
	api := router.Group("/api")
	{
		// ── Public endpoints (no token required) ──
		api.POST("/auth/send-otp", authCtrl.SendOTP)
		api.POST("/signup", farmerCtrl.Signup)
		api.POST("/login", farmerCtrl.Login)

		// ── Protected endpoints (JWT token required) ──
		protected := api.Group("")
		protected.Use(middlewares.JWTAuth(jwtService))
		{
			protected.POST("/logout", farmerCtrl.Logout)
			protected.PUT("/location", farmerCtrl.UpdateLocation)
			protected.POST("/soil/upload", soilCtrl.UploadSoil)
			protected.POST("/chat", chatCtrl.Chat)
			protected.GET("/weather", weatherCtrl.GetWeather)
			protected.POST("/samyakai", samyakAICtrl.Chat)
			protected.POST("/voice/tts", voiceCtrl.TextToSpeech)
		}
	}

	// Health check (always public)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "SamyakSetu API",
		})
	})
}
