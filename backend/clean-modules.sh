#!/bin/bash
set -e  # Exit immediately if a command fails

echo "Cleaning Go module cache..."
go clean -modcache

echo "Removing go.sum..."
rm -f go.sum

echo "Tidying Go modules..."
go mod tidy

echo "Done! Your go.sum file has been regenerated."
