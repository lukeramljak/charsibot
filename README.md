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

2. Install dependencies for all services:

   ```bash
   # Discord bot
   cd discord && bun install

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

### Development

Run all services:

```bash
bun run dev
```

Run individual services:

```bash
bun run dev:discord
bun run dev:twitch
bun run dev:overlay
```

### Production (Docker Compose)

**Run all services:**

```bash
docker compose up -d
```

**Run individual service:**

```bash
docker compose up discord -d

docker compose up twitch -d

# Overlay only
docker compose up twitch-overlay -d
```

**Build and restart:**

```bash
docker compose up --build -d
```
