#!/bin/bash

# ==============================================================================
# SamyakSetu Auto-Deployment Script
# Run this script anytime you make changes to the Go source code to automatically
# deploy the new version straight to the AWS EC2 server!
# ==============================================================================

SERVER_IP="51.21.199.205"
PEM_KEY="/home/devrishijain/Documents/Personal/samyak-mongo-key.pem"
APP_NAME="samyak-backend"

echo "🚀 Starting Deployment to AWS ($SERVER_IP)..."

echo "📦 1/3 Compiling Go Code for Ubuntu Server..."
GOOS=linux GOARCH=amd64 go build -o $APP_NAME cmd/main.go
if [ $? -ne 0 ]; then
    echo "❌ Build failed. Please fix your Go errors."
    exit 1
fi

echo "📤 2/3 Uploading new version and variables to EC2..."
scp -i "$PEM_KEY" -o StrictHostKeyChecking=no $APP_NAME .env ubuntu@$SERVER_IP:/home/ubuntu/
if [ $? -ne 0 ]; then
    echo "❌ Upload failed. Ensure your PEM key path is correct and server is running."
    exit 1
fi

echo "🔄 3/3 Restarting exactly running background service on AWS..."
ssh -i "$PEM_KEY" -o StrictHostKeyChecking=no ubuntu@$SERVER_IP "sudo systemctl restart samyak && sudo systemctl status samyak --no-pager"

echo "🧹 Cleaning up local files..."
rm -f $APP_NAME

echo "✅ SUCCESS! The new backend is fully live at http://$SERVER_IP:8080"
