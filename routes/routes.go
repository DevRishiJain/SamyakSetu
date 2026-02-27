// All rights reserved Samyak-Setu

package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/controllers"
)

// RegisterRoutes sets up all API routes on the Gin engine.
func RegisterRoutes(
	router *gin.Engine,
	authCtrl *controllers.AuthController,
	farmerCtrl *controllers.FarmerController,
	soilCtrl *controllers.SoilController,
	chatCtrl *controllers.ChatController,
) {
	api := router.Group("/api")
	{
		// Auth endpoints
		api.POST("/auth/send-otp", authCtrl.SendOTP)

		// Farmer endpoints
		api.POST("/signup", farmerCtrl.Signup)
		api.PUT("/location", farmerCtrl.UpdateLocation)

		// Soil endpoints
		api.POST("/soil/upload", soilCtrl.UploadSoil)

		// Chat endpoints
		api.POST("/chat", chatCtrl.Chat)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "SamyakSetu API",
		})
	})
}
