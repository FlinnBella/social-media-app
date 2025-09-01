#!/bin/bash
set -e  # Exit immediately if a command fails

export APP_ENV=production

echo "Starting production server..."

echo "Environment: $APP_ENV"

go run main.go

echo "Production server started"