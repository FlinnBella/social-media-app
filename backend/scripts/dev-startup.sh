#!/bin/bash
set -e
export APP_ENV=development

cd "$(dirname "$0")"  # ensure we're in backend/
echo "Starting development server..."
echo "Environment: $APP_ENV"


# bind loopback explicitly (no delve debug logs)
dlv debug . --headless --listen=127.0.0.1:40000 --api-version=2 --accept-multiclient