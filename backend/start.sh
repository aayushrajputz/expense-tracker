#!/bin/bash

# Production startup script for Expense Tracker Backend

echo "Starting Expense Tracker Backend..."

# Set default values if not provided
export PORT=${PORT:-8080}
export APP_ENV=${APP_ENV:-production}

echo "Configuration:"
echo "  PORT: $PORT"
echo "  APP_ENV: $APP_ENV"
echo "  DB_DSN: ${DB_DSN:+***set***}"

# Check if required environment variables are set
if [ -z "$JWT_SECRET" ]; then
    echo "ERROR: JWT_SECRET environment variable is required"
    exit 1
fi

if [ -z "$DB_DSN" ]; then
    echo "ERROR: DB_DSN environment variable is required"
    exit 1
fi

# Run the application
echo "Starting application on port $PORT..."
exec ./main 