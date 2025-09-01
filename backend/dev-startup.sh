#!/bin/bash
set -e  # Exit immediately if a command fails

export APP_ENV=development

echo "Starting development server..."

echo "Environment: $APP_ENV"

go run main.go

echo "Development server started"