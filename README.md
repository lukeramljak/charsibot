# charsibot

A multi-service bot application for [Charsibel](https://twitch.tv/charsibel)'s Discord server and Twitch channel.

## Architecture

This project contains two independent services:

- **Discord Bot** (`/discord`) - Handles Discord server interactions and commands
- **Twitch Bot** (`/twitch`) - Manages Twitch chat commands and channel point redemptions

## Prerequisites

- Go 1.25+
- Bun
- Docker & Docker Compose

## Development Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/lukeramljak/charsibot.git
   cd charsibot
   ```

2. Install dependencies for both services:

   ```bash
   cd discord && go mod tidy
   cd twitch && bun install
   ```

3. Set up environment variables:

   ```bash
   # Discord bot
   cp discord/.env.example discord/.env
   # Edit discord/.env with your Discord credentials

   # Twitch bot
   cp twitch/.env.example twitch/.env
   # Edit twitch/.env with your Twitch credentials
   ```

## Running Services

### Development (Individual Services)

**Discord Bot:**

```bash
cd discord
air main.go
```

**Twitch Bot:**

```bash
cd twitch
bun run dev
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
