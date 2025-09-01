#!/bin/bash

set -e

echo "Deploying Charsibot Services..."

# Function to deploy specific service
deploy_service() {
    local service=$1
    echo "Deploying $service..."
    docker compose up $service --build -d
}

# Function to deploy all services
deploy_all() {
    echo "Deploying all services..."
    docker compose up --build -d
}

# Parse command line arguments
case "${1:-all}" in
    "discord")
        deploy_service "charsibot-discord"
        ;;
    "twitch")
        deploy_service "charsibot-twitch"
        ;;
    "all"|"")
        deploy_all
        ;;
    *)
        echo "Usage: $0 [discord|twitch|all]"
        echo "  discord - Deploy only Discord bot"
        echo "  twitch  - Deploy only Twitch bot"
        echo "  all     - Deploy both services (default)"
        exit 1
        ;;
esac

echo ""
echo "Container status:"
docker compose ps

echo ""
echo "Useful commands:"
echo "  View logs (all):     docker compose logs -f"
echo "  View logs (discord): docker compose logs -f charsibot-discord"
echo "  View logs (twitch):  docker compose logs -f charsibot-twitch"
echo "  Stop all services:   docker compose down"
echo ""
echo "Deployment complete! 🎉"
