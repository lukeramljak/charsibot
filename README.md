# charsibot

A multi-service bot application for [Charsibel](https://twitch.tv/charsibel)'s Discord server and Twitch channel.

## Architecture

This project contains three services:

- **Discord Bot** (`discord/`) - Handles Discord server interactions and commands (Node.js/Bun)
- **Twitch Bot** (`twitch/`) - Manages Twitch chat commands, channel point redemptions, and API server (Go)
- **Twitch Overlay** (`twitch-overlay/`) - Stream overlay SPA for OBS (SvelteKit)

## Prerequisites

- Bun
- Go 1.25+
- Docker

## Development Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/lukeramljak/charsibot.git
   cd charsibot
   ```

2. **Discord Bot**:

   ```bash
   cd discord
   bun install
   bun run dev
   ```

3. **Twitch Bot**:

   ```bash
   cd twitch
   make dev
   ```

   This will start the Go backend and API server on port 8081.

4. **Twitch Overlay**:

   ```bash
   cd twitch-overlay
   bun install
   bun run dev
   ```

   Access the overlay at `http://localhost:5173`

## Environment Variables

1. **Discord**:

   ```bash
   cp discord/.env.example discord/.env
   ```

2. **Twitch**:
   ```bash
   cp twitch/.env.example twitch/.env
   ```

## Database

The Twitch bot uses a SQLite database (`charsibot.db`) which will be created automatically in the root of the `twitch` directory.
