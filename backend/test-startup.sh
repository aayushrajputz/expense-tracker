#!/bin/bash

# Test script to verify the application can start properly

echo "Testing application startup..."

# Set required environment variables
export PORT=8080
export APP_ENV=development
export DB_DSN="postgres://postgres:postgres@localhost:5432/expensedb?sslmode=disable"
export JWT_SECRET="test-secret-key"

# Build the application
echo "Building application..."
go build -o test-app ./cmd/server

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful!"

# Test if the binary can start (timeout after 10 seconds)
echo "Testing application startup..."
timeout 10s ./test-app &
APP_PID=$!

# Wait a moment for the app to start
sleep 3

# Check if the process is still running
if kill -0 $APP_PID 2>/dev/null; then
    echo "Application started successfully!"
    
    # Test health endpoint
    echo "Testing health endpoint..."
    curl -f http://localhost:8080/health
    if [ $? -eq 0 ]; then
        echo "Health endpoint working!"
    else
        echo "Health endpoint failed!"
    fi
    
    # Test ping endpoint
    echo "Testing ping endpoint..."
    curl -f http://localhost:8080/ping
    if [ $? -eq 0 ]; then
        echo "Ping endpoint working!"
    else
        echo "Ping endpoint failed!"
    fi
    
    # Kill the application
    kill $APP_PID
    wait $APP_PID 2>/dev/null
else
    echo "Application failed to start!"
    exit 1
fi

# Clean up
rm -f test-app

echo "Test completed successfully!" 