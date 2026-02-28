# SamyakSetu AI — Backend Architecture & API Documentation

Welcome to the backend repository of **SamyakSetu**, a production-grade Agricultural AI platform built in Golang. 

This document serves as an end-to-end guide for frontend developers, mobile app engineers, and new contributors to understand the system architecture, how the APIs talk to each other, and exactly how to consume every endpoint.

---

## 🏗️ Architecture Overview

The backend was engineered using **Clean Architecture** principles to ensure that external services (like AI and Databases) can be hot-swapped without breaking the core business logic.

1. **Language & Framework:** Golang 1.23+ with the `Gin` HTTP framework.
2. **Database (MongoDB on AWS EC2):** Used for storing Farmers (`farmers`), Soil Image data (`soil_data`), Chat Histories (`chat_messages`), and ephemeral data (`otp_codes`).
3. **AI Brain (Amazon Nova Lite via AWS Bedrock):** 
   - **Vision Model:** Reads uploaded soil images to accurately detect the soil type (e.g., Clay, Loamy, Alluvial).
   - **Text Model:** Powers the advisory chat, injecting live weather, GPS data, and soil type context into the LLM prompt.
4. **Cloud Storage (AWS S3):** Uploaded soil images are streamed directly to a public Amazon S3 Bucket to securely scale storage without bloating the Golang server.
5. **Real-time Weather (OpenWeatherMap API):** Grabs real-time weather metadata based on the farmer's GPS coordinates to enrich the AI's agricultural advice safely.
6. **Authentication:** JWT-based session tokens with Prototype Mode (master OTP `000000`) for easy demo/testing.

---

## 🔐 Authentication Flow

SamyakSetu uses **JWT (JSON Web Tokens)** for session management. Here is the flow:

1. **Signup/Login** → Server returns a `token` field in the response.
2. **Store the token** on the frontend (localStorage, SecureStorage, etc.).
3. **Send the token with every request** in the `Authorization` header:
   ```
   Authorization: Bearer <your_token_here>
   ```
4. **Logout** → Frontend clears the stored token.

### 🧪 Prototype Mode
Since this is a prototype, a real SMS gateway is not connected yet. Instead, **Prototype Mode** is enabled:
- You can use the master OTP **`000000`** during signup and login without calling `/send-otp` first.
- This allows the frontend team to create accounts and test freely without needing the backend console.
- To disable this for production, set `PROTOTYPE_MODE=false` in the `.env` file.

---

## 📡 End-to-End API Documentation

Below are all the endpoints exposed for the frontend.

**Base URLs:**
- Local Development: `http://localhost:8080`
- Production (EC2): `http://51.21.199.205:8080`

---

### 1. Health Check
Always hit this to ensure the server is alive.

- **Endpoint**: `GET /health`
- **Auth Required**: ❌ No
- **cURL Example**:
  ```bash
  curl -X GET http://51.21.199.205:8080/health
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "service": "SamyakSetu API",
      "status": "ok"
  }
  ```

---

### 2. Request OTP (Registration Step 1 — Optional in Prototype Mode)
Sends a strict 6-digit verification code to the farmer's phone number before they can register.

> **Note:** In Prototype Mode, you can skip this step entirely and use `"otp": "000000"` during signup/login.

- **Endpoint**: `POST /api/auth/send-otp`
- **Auth Required**: ❌ No
- **Content-Type**: `application/json`
- **cURL Example**:
  ```bash
  curl -X POST http://51.21.199.205:8080/api/auth/send-otp \
    -H "Content-Type: application/json" \
    -d '{"phone": "9988776655"}'
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "message": "OTP sent successfully. It will expire in 5 minutes."
  }
  ```

---

### 3. Signup Farmer (Registration Step 2)
Registers the farmer and returns a **JWT token** for the current session.

- **Endpoint**: `POST /api/signup`
- **Auth Required**: ❌ No
- **Content-Type**: `application/json`
- **cURL Example (Prototype Mode — using master OTP)**:
  ```bash
  curl -X POST http://51.21.199.205:8080/api/signup \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Rajesh Kumar",
      "phone": "9988776655",
      "otp": "000000",
      "latitude": 28.6139,
      "longitude": 77.2090
    }'
  ```
- **Success Response** (`201 Created`):
  ```json
  {
      "id": "69a2f4726f2bd4aa38a6314f",
      "name": "Rajesh Kumar",
      "phone": "9988776655",
      "location": {
          "latitude": 28.6139,
          "longitude": 77.209
      },
      "createdAt": "2026-02-28T19:28:10Z",
      "token": "eyJhbGciOiJIUzI1NiIs..."
  }
  ```
  > ⚠️ **IMPORTANT:** Save the `id` and `token` from this response. The `token` must be sent in the `Authorization` header for all subsequent API calls. The `id` is the `farmerId` needed for chat, soil, and location APIs.

---

### 4. Login (Returning Users)
Authenticates an existing farmer by phone number and OTP, and returns a fresh **JWT token**.

- **Endpoint**: `POST /api/login`
- **Auth Required**: ❌ No
- **Content-Type**: `application/json`
- **cURL Example (Prototype Mode)**:
  ```bash
  curl -X POST http://51.21.199.205:8080/api/login \
    -H "Content-Type: application/json" \
    -d '{
      "phone": "9988776655",
      "otp": "000000"
    }'
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "id": "69a2f4726f2bd4aa38a6314f",
      "name": "Rajesh Kumar",
      "phone": "9988776655",
      "location": {
          "latitude": 28.6139,
          "longitude": 77.209
      },
      "token": "eyJhbGciOiJIUzI1NiIs..."
  }
  ```

---

### 5. Logout
Logs out the current session. The frontend should also clear the stored token.

- **Endpoint**: `POST /api/logout`
- **Auth Required**: ✅ Yes (`Authorization: Bearer <token>`)
- **cURL Example**:
  ```bash
  curl -X POST http://51.21.199.205:8080/api/logout \
    -H "Authorization: Bearer YOUR_TOKEN_HERE"
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "message": "Logged out successfully"
  }
  ```

---

### 6. Upload Soil Image & Get AI Analysis
Uploads an image from the farmer's camera, stores it in AWS S3, and sends it to the Amazon Nova Lite Vision model to detect exactly what kind of soil it is.

- **Endpoint**: `POST /api/soil/upload`
- **Auth Required**: ✅ Yes (`Authorization: Bearer <token>`)
- **Content-Type**: `multipart/form-data`
- **Parameters**:
  - `farmerId` (string): The ObjectID of the registered farmer.
  - `soilImage` (file): The actual image file (JPEG, PNG, WebP, or GIF).
- **cURL Example**:
  ```bash
  curl -X POST http://51.21.199.205:8080/api/soil/upload \
    -H "Authorization: Bearer YOUR_TOKEN_HERE" \
    -F "farmerId=69a2f4726f2bd4aa38a6314f" \
    -F "soilImage=@/path/to/my_soil.jpg"
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "soilType": "Alluvial Soil",
      "imagePath": "https://samyak-setu-soil.s3.eu-north-1.amazonaws.com/soil/123456789.jpg"
  }
  ```

---

### 7. Talk to Agronomy AI Advisor (Chat)
Sends a natural language question (and an optional image) to the agricultural AI. **The backend automatically enriches this prompt** by pulling the farmer's GPS coordinates, live weather forecasts, and newest soil analysis from MongoDB!

- **Endpoint**: `POST /api/chat`
- **Auth Required**: ✅ Yes (`Authorization: Bearer <token>`)
- **Content-Type**: Can be `application/json` (text-only) OR `multipart/form-data` (text + image attachment).
- **Parameters**:
  - `farmerId` (string): The ObjectID of the registered farmer.
  - `message` (string): The question asked by the farmer.
  - `image` (file, optional): An image to help the AI understand pest/crop diseases.
- **cURL Example (Text Only - JSON)**:
  ```bash
  curl -X POST http://51.21.199.205:8080/api/chat \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer YOUR_TOKEN_HERE" \
    -d '{
      "farmerId": "69a2f4726f2bd4aa38a6314f",
      "message": "My crop leaves are turning yellow, what should I do?"
    }'
  ```
- **cURL Example (With Image - Multipart)**:
  ```bash
  curl -X POST http://51.21.199.205:8080/api/chat \
    -H "Authorization: Bearer YOUR_TOKEN_HERE" \
    -F "farmerId=69a2f4726f2bd4aa38a6314f" \
    -F "message=What is eating my tomatoes?" \
    -F "image=@/path/to/tomato_bug.jpg"
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "reply": "I see you are dealing with yellowing leaves on your crops near Surat where it's currently 35°C. Since you have Loamy Soil, this is highly likely a nitrogen deficiency..."
  }
  ```

---

### 8. Update Farmer Location
Quietly updates the GPS coordinates of the farmer in MongoDB (useful if the frontend polls GPS dynamically as the farmer walks their field).

- **Endpoint**: `PUT /api/location`
- **Auth Required**: ✅ Yes (`Authorization: Bearer <token>`)
- **Content-Type**: `application/json`
- **cURL Example**:
  ```bash
  curl -X PUT http://51.21.199.205:8080/api/location \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer YOUR_TOKEN_HERE" \
    -d '{
      "farmerId": "69a2f4726f2bd4aa38a6314f",
      "latitude": 26.9124,
      "longitude": 75.7873
    }'
  ```
- **Success Response** (`200 OK`):
  ```json
  {
      "farmerId": "69a2f4726f2bd4aa38a6314f",
      "location": {
          "latitude": 26.9124,
          "longitude": 75.7873
      },
      "message": "Location updated successfully"
  }
  ```

---

## ❌ Error Responses

All error responses follow the same format:
```json
{
    "error": "Description of what went wrong"
}
```

| Status Code | Meaning |
|-------------|---------|
| `400` | Bad Request — invalid input, missing fields, bad phone format |
| `401` | Unauthorized — missing/invalid/expired JWT token, or bad OTP |
| `404` | Not Found — farmer or resource doesn't exist |
| `409` | Conflict — phone number already registered |
| `500` | Internal Server Error — something broke on the server |

---

## 🔒 Local Setup for Backend Developers

1. **Clone Repo & Install Modules**
   ```bash
   go mod tidy
   ```
2. **Setup Local Environment Variables**
   Duplicate `.env.example` as `.env`, and provide real API keys for:
   - MongoDB URI
   - AWS Bedrock Access Key + Secret Key
   - OpenWeatherMap API Key
   - AWS S3 Region + Keys + Bucket Name
   - `PROTOTYPE_MODE=true` (for development)
   - `JWT_SECRET` (any random long string)
3. **Start the MongoDB Service**
   Ensure `mongod` is running on your machine or connect to the EC2 instance.
4. **Compile & Run**
   ```bash
   go build -o app ./cmd/main.go && ./app
   ```
5. **Deploy to EC2**
   ```bash
   ./deploy.sh
   ```