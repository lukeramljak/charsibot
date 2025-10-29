# charsibot

A multi-service bot application for [Charsibel](https://twitch.tv/charsibel)'s Discord server and Twitch channel.

## Architecture

This project contains three services:

- **Discord Bot** (`/discord`) - Handles Discord server interactions and commands
- **Twitch Bot** (`/twitch`) - Manages Twitch chat commands and channel point redemptions
- **Twitch Overlay** (`/twitch-overlay`) - SvelteKit web app providing browser source overlays for OBS

## Prerequisites

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
   # Discord bot
   cd discord && go mod tidy

   # Twitch bot
   cd twitch && bun install

   # Twitch overlay
   cd twitch-overlay && bun install
   ```

3. Set up environment variables:

   ```bash
   # Discord bot
   cp discord/.env.example discord/.env
   # Edit discord/.env with your Discord credentials

   # Twitch bot
   cp twitch/.env.example twitch/.env
   # Edit twitch/.env with your Twitch credentials

   # Twitch overlay
   cp twitch-overlay/.env.example twitch-overlay/.env
   # Edit twitch-overlay/.env with WebSocket URL
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

**Twitch Overlay:**

```bash
cd twitch-overlay
bun run dev
```

### Production (Docker Compose)

**Run all services:**

```bash
docker compose up -d
```

**Run individual service:**

```bash
# Discord only
docker compose up charsibot-discord -d

# Twitch only
docker compose up charsibot-twitch -d

# Overlay only
docker compose up twitch-overlay -d
```

**Build and restart:**

```bash
docker compose up --build -d
```
