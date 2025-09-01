# charsibot

A multi-service bot application for [Charsibel](https://twitch.tv/charsibel)'s Discord server and Twitch channel.

## Architecture

This project contains two independent services:

- **Discord Bot** (`/discord`) - Handles Discord server interactions and commands
- **Twitch Bot** (`/twitch`) - Manages Twitch chat commands and channel point redemptions

## Prerequisites

- Go 1.25+
- Docker & Docker Compose

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/lukeramljak/charsibot.git
   cd charsibot
   ```

2. Install dependencies for both services:
   ```bash
   go work sync
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your actual credentials
   ```

## Running Services

### Development (Individual Services)

**Discord Bot:**
```bash
cd discord
go run ./cmd/bot
```

**Twitch Bot:**
```bash
cd twitch
go run .
```

### Production (Docker Compose)

**Run both services:**
```bash
docker compose up -d
```

**Run individual service:**
```bash
# Discord only
docker compose up charsibot-discord -d

# Twitch only
docker compose up charsibot-twitch -d
```

**Build and restart:**
```bash
docker compose up --build -d
```

## Deployment

### Option 1: Root Compose (Recommended)
```bash
git pull
docker compose up --build -d
```

### Option 2: Individual Services
Navigate to the specific service directory and use its individual `docker-compose.yml`:
```bash
cd discord  # or twitch
docker compose up --build -d
```

## Service Details

### Discord Bot (`/discord`)
- **Purpose**: Handles Discord server interactions and commands for Charsibel's server
- **Technology**: Go 1.25, DiscordGo library
- **Features**: Custom commands, event handling, message processing

### Twitch Bot (`/twitch`)
- **Purpose**: Manages Twitch chat commands and channel point redemptions
- **Technology**: Go 1.25, WebSocket connections, Turso database
- **Features**: Chat commands, channel points, websocket connections, logging
