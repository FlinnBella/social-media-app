#!/bin/bash
set -e
export APP_ENV=development

cd "$(dirname "$0")"  # ensure we're in backend/
echo "Starting development server..."
echo "Environment: $APP_ENV"
