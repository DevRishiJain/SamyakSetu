# SamyakSetu Backend

Agricultural AI advisory platform backend built with Go, Gin, and MongoDB.

## Features

- Farmer signup with location tracking
- Soil image upload with AI-powered analysis (Google Gemini Vision)
- AI advisory chat with weather + soil context
- Weather integration (OpenWeatherMap)
- Clean Architecture with interface-driven design
- Production-ready with graceful shutdown

## Quick Start

```bash
# 1. Copy environment file
cp .env.example .env
# Edit .env with your API keys

# 2. Install dependencies
go mod tidy

# 3. Run the server
go run cmd/main.go

# Or build and run
go build -o app cmd/main.go && ./app
```

## API Endpoints

| Method | Path              | Description                    |
|--------|-------------------|--------------------------------|
| POST   | /api/signup       | Register a new farmer          |
| PUT    | /api/location     | Update farmer GPS location     |
| POST   | /api/soil/upload  | Upload soil image for analysis |
| POST   | /api/chat         | AI-powered advisory chat       |

## Architecture

```
cmd/            → Application entry point
config/         → Environment configuration
database/       → MongoDB connection manager
models/         → Data models
repositories/   → Database access layer
services/       → Business logic + external API integrations
controllers/    → HTTP request handlers
routes/         → Route definitions
middlewares/    → CORS, logging, etc.
utils/          → Validators and helpers
```

## Environment Variables

| Variable        | Description                    |
|-----------------|--------------------------------|
| PORT            | Server port (default: 8080)    |
| MONGO_URI       | MongoDB connection string      |
| GEMINI_API_KEY  | Google Gemini API key          |
| WEATHER_API_KEY | OpenWeatherMap API key         |
| UPLOAD_PATH     | File upload directory          |

## License

MIT