# SamyakSetu AI — Backend Architecture & API Documentation

Welcome to the backend repository of **SamyakSetu**, a production-grade Agricultural AI platform built in Golang. 

This document serves as an end-to-end guide for frontend developers, mobile app engineers, and new contributors to understand the system architecture, how the APIs talk to each other, and exactly how to consume every endpoint.

---

## 🏗️ Architecture Overview

The backend was engineered using **Clean Architecture** principles to ensure that external services (like AI and Databases) can be hot-swapped without breaking the core business logic.

1. **Language & Framework:** Golang 1.23+ with the `Gin` HTTP framework.
2. **Database (MongoDB):** Used for storing Farmers (`farmers`), Soil Image data (`soil_data`), Chat Histories (`chat_messages`), and ephemeral data (`otp_codes`).
3. **AI Brain (Google Gemini 2.0 Flash):** 
   - **Vision Model:** Reads uploaded soil images to accurately detect the soil type (e.g., Clay, Loamy, Alluvial).
   - **Text Model:** Powers the advisory chat, injecting live weather, GPS data, and soil type context into the LLM prompt.
4. **Cloud Storage (AWS S3):** Uploaded soil images are streamed directly to a public Amazon S3 Bucket to securely scale storage without bloating the Golang server.
5. **Real-time Weather (OpenWeatherMap API):** Grabs real-time weather metadata based on the farmer's GPS coordinates to enrich the AI's agricultural advice safely.
6. **OTP Authentication Service:** A mockable OTP flow that issues cryptographically secure 6-digit codes and saves them to Mongo with a strict 5-minute Time-to-Live (TTL).

---

## 📡 End-to-End API Documentation

Below are all the endpoints exposed for the frontend. 

**Base URLs:**
- Local Development: `http://localhost:8080/api`
- Production (EC2): `http://51.21.199.205:8080/api`

### 1. Health Check
Always hit this to ensure the server is alive.

- **Endpoint**: `GET /health`
- **Request Parameters**: None
- **cURL Example**:
  ```bash
  curl -X GET http://localhost:8080/health
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "service": "SamyakSetu API",
      "status": "ok"
  }
  ```

---

### 2. Request OTP (Registration Step 1)
Sends a strict 6-digit verification code to the farmer's phone number before they can register.

- **Endpoint**: `POST /api/auth/send-otp`
- **Content-Type**: `application/json`
- **cURL Example**:
  ```bash
  curl -X POST http://localhost:8080/api/auth/send-otp \
    -H "Content-Type: application/json" \
    -d '{"phone": "9988776655"}'
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "message": "OTP sent successfully. It will expire in 5 minutes."
  }
  ```
  *(Note for Devs: In development mode, the OTP is printed directly in the backend terminal console so you don't need real SMS delivery).*

---

### 3. Verify OTP & Signup Farmer (Registration Step 2)
Registers the farmer into MongoDB securely, strictly enforcing that their OTP is correct and hasn't expired.

- **Endpoint**: `POST /api/signup`
- **Content-Type**: `application/json`
- **cURL Example**:
  ```bash
  curl -X POST http://localhost:8080/api/signup \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Rajesh Kumar",
      "phone": "9988776655",
      "otp": "540993",
      "latitude": 28.6139,
      "longitude": 77.2090
    }'
  ```
- **Success Response** (`201 Created`):
  ```json
  {
      "id": "69a1ede5964d732585013ca2",
      "name": "Rajesh Kumar",
      "phone": "9988776655",
      "location": {
          "latitude": 28.6139,
          "longitude": 77.209
      },
      "createdAt": "2026-02-28T00:47:57.918Z"
  }
  ```
  *(Keep this `id` safe on the frontend—you will need the `farmerId` for all subsequent API calls).*

---

### 4. Upload Soil Image & Get AI Analysis
Uploads an image from the farmer's camera, stores it in AWS S3, and sends it to the Gemini Vision model to detect exactly what kind of soil it is.

- **Endpoint**: `POST /api/soil/upload`
- **Content-Type**: `multipart/form-data`
- **Parameters**:
  - `farmerId` (string): The ObjectID of the registered farmer.
  - `soilImage` (file): The actual image file.
- **cURL Example**:
  ```bash
  curl -X POST http://localhost:8080/api/soil/upload \
    -F "farmerId=69a1ede5964d732585013ca2" \
    -F "soilImage=@/path/to/my_soil.jpg"
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "soilType": "Alluvial Soil",
      "imagePath": "https://samyak-setu-soil.s3.eu-north-1.amazonaws.com/soil/1772219353755488768.jpg"
  }
  ```
  *(Note: The AI gracefully falls back to "Unknown (Pending AI Analysis)" if Google limits are reached, but the S3 upload successfully runs anyway).*

---

### 5. Talk to Agronomy AI Advisor (Chat)
Sends a natural language question (and an optional image) to the agricultural AI. **The backend automatically enriches this prompt** by pulling the farmer's GPS coordinates, live weather forecasts, and newest soil analysis from MongoDB! 

- **Endpoint**: `POST /api/chat`
- **Content-Type**: Can be `application/json` (text-only) OR `multipart/form-data` (text + image attachment).
- **Parameters**:
  - `farmerId` (string): The ObjectID of the registered farmer.
  - `message` (string): The question asked by the farmer.
  - `image` (file, optional): An image to help the AI understand pest/crop diseases.
- **cURL Example (Text Only - JSON)**:
  ```bash
  curl -X POST http://localhost:8080/api/chat \
    -H "Content-Type: application/json" \
    -d '{
      "farmerId": "69a1ede5964d732585013ca2",
      "message": "My crop leaves are turning yellow, what should I do?"
    }'
  ```
- **cURL Example (With Image - Multipart)**:
  ```bash
  curl -X POST http://localhost:8080/api/chat \
    -F "farmerId=69a1ede5964d732585013ca2" \
    -F "message=What is eating my tomatoes?" \
    -F "image=@/path/to/tomato_bug.jpg"
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "reply": "I see you are dealing with yellowing leaves on your crops in Jaipur where it's currently 35°C. Since you have Red Soil, this is highly likely an iron deficiency exacerbated by the heat..."
  }
  ```

---

### 6. Update Farmer Location
Quietly updates the GPS coordinates of the farmer in MongoDB (useful if the frontend polls GPS dynamically as the farmer walks their field).

- **Endpoint**: `PUT /api/location`
- **Content-Type**: `application/json`
- **cURL Example**:
  ```bash
  curl -X PUT http://localhost:8080/api/location \
    -H "Content-Type: application/json" \
    -d '{
      "farmerId": "69a1ede5964d732585013ca2",
      "latitude": 26.9124,
      "longitude": 75.7873
    }'
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "farmerId": "69a1ede5964d732585013ca2",
      "location": {
          "latitude": 26.9124,
          "longitude": 75.7873
      },
      "message": "Location updated successfully"
  }
  ```

---

## 🔒 Local Setup for Backend Developers

1. **Clone Repo & Install Modules**
   ```bash
   go mod tidy
   ```
2. **Setup Local Environment Variables**
   Duplicate `.env.example` as `.env`, and provide real API keys for:
   - MongoDB URI
   - Google Gemini API Key
   - OpenWeatherMap API Key
   - AWS S3 Region + Keys + Bucket Name
3. **Start the MongoDB Service**
   Ensure `mongod` is running on your machine securely.
4. **Compile & Run**
   ```bash
   go build -o app ./cmd/main.go && ./app
   ```