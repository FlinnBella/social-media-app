#!/bin/bash
set -e  # Exit immediately if a command fails

export APP_ENV=production

echo "Starting production server..."

cd "$(dirname "$0")"  # ensure we're in backend/


echo "Environment: $APP_ENV"
# bind loopback explicitly (no delve debug logs)
dlv debug . --headless --listen=127.0.0.1:40000 --api-version=2 --accept-multiclient --log --log-output=debugger,stdout,stderr

echo "Production server started"