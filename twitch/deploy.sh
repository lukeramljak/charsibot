#!/bin/bash

set -e

echo "Deploying Charsibot Twitch Bot..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ .env file not found. Please copy .env.example to .env and fill in your credentials."
    echo "   cp .env.example .env"
    exit 1
fi

echo "Stopping existing container..."
docker-compose down || true

echo "Building and starting container..."
docker-compose up --build -d

echo "Container status:"
docker-compose ps

echo "To view logs:"
echo "   docker-compose logs -f"
echo ""
echo "To stop the bot:"
echo "   docker-compose down"
echo ""
echo "Deployment complete!"
