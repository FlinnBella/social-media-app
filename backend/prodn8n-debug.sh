#!/bin/bash
set -e  # Exit immediately if a command fails

export APP_ENV=production

echo "Starting production server..."

echo "Environment: $APP_ENV"
# bind loopback explicitly (no delve debug logs)
dlv debug . --headless --listen=127.0.0.1:40000 --api-version=2 --accept-multiclient

echo "Production server started"