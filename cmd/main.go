// All rights reserved Samyak-Setu

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/config"
	"github.com/samyaksetu/backend/controllers"
	"github.com/samyaksetu/backend/database"
	"github.com/samyaksetu/backend/middlewares"
	"github.com/samyaksetu/backend/repositories"
	"github.com/samyaksetu/backend/routes"
	"github.com/samyaksetu/backend/services"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("INFO: Starting SamyakSetu Backend...")

	// Load configuration
	cfg := config.LoadConfig()

	// Connect to MongoDB
	db, err := database.Connect(cfg.MongoURI)
	if err != nil {
		log.Fatalf("FATAL: MongoDB connection failed: %v", err)
	}
	defer db.Disconnect()

	// Initialize AI service
	var aiService services.AIService
	if cfg.BedrockAccessKey != "" && cfg.BedrockSecretKey != "" {
		aiService, err = services.NewBedrockService(cfg.BedrockRegion, cfg.BedrockAccessKey, cfg.BedrockSecretKey, cfg.BedrockSessionToken)
		if err != nil {
			log.Fatalf("FATAL: Bedrock service initialization failed: %v", err)
		}
	} else if cfg.GeminiAPIKey != "" {
		geminiService, err := services.NewGeminiService(cfg.GeminiAPIKey)
		if err != nil {
			log.Fatalf("FATAL: Gemini service initialization failed: %v", err)
		}
		defer geminiService.Close()
		aiService = geminiService
	} else {
		log.Fatalf("FATAL: No AI service configured. Set GEMINI_API_KEY or BEDROCK credentials.")
	}

	weatherService := services.NewOpenWeatherService(cfg.WeatherAPIKey)

	var storageService services.StorageService
	if cfg.S3BucketName != "" {
		storageService, err = services.NewS3StorageService(cfg.AWSRegion, cfg.AWSAccessKey, cfg.AWSSecretKey, cfg.S3BucketName)
		if err != nil {
			log.Fatalf("FATAL: AWS S3 service initialization failed: %v", err)
		}
	} else {
		storageService, err = services.NewLocalStorageService(cfg.UploadPath)
		if err != nil {
			log.Fatalf("FATAL: Storage service initialization failed: %v", err)
		}
	}

	farmerRepo := repositories.NewFarmerRepository(db)
	soilRepo := repositories.NewSoilRepository(db)
	chatRepo := repositories.NewChatRepository(db)
	otpRepo := repositories.NewOTPRepository(db)

	// Initialize JWT service for session management
	jwtService := services.NewJWTService(cfg.JWTSecret)
	if cfg.PrototypeMode {
		log.Println("INFO: 🧪 PROTOTYPE MODE is ON — master OTP \"000000\" can be used to bypass OTP verification")
	}

	// Initialize controllers (dependency injection)
	authCtrl := controllers.NewAuthController(otpRepo, services.NewMockOTPService())
	farmerCtrl := controllers.NewFarmerController(farmerRepo, otpRepo, jwtService, cfg.PrototypeMode)
	soilCtrl := controllers.NewSoilController(farmerRepo, soilRepo, aiService, storageService)
	chatCtrl := controllers.NewChatController(farmerRepo, soilRepo, chatRepo, aiService, weatherService)
	weatherCtrl := controllers.NewWeatherController(farmerRepo, weatherService)
	samyakAICtrl := controllers.NewSamyakAIController(aiService)

	var voiceService services.VoiceService
	voiceService, err = services.NewAWSVoiceService(cfg.AWSRegion, cfg.AWSAccessKey, cfg.AWSSecretKey, cfg.S3BucketName, storageService)
	if err != nil {
		log.Printf("WARN: Failed to initialize AWS Voice Service: %v (TTS might not work)", err)
	}
	voiceCtrl := controllers.NewVoiceController(voiceService, aiService)

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middlewares.SetupCORS())
	router.Use(middlewares.RequestLogger())

	// Register routes
	routes.RegisterRoutes(router, authCtrl, farmerCtrl, soilCtrl, chatCtrl, weatherCtrl, samyakAICtrl, voiceCtrl, jwtService)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("INFO: Server listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("FATAL: Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("INFO: Received signal %v — shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("ERROR: Server forced shutdown: %v", err)
	}

	log.Println("INFO: SamyakSetu Backend stopped")
}
