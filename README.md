# charsibot

[![Go Version](https://img.shields.io/github/go-mod/go-version/lukeramljak/charsibot?filename=twitch%2Fgo.mod)](https://github.com/lukeramljak/charsibot/blob/main/twitch/go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/lukeramljak/charsibot/twitch)](https://goreportcard.com/report/github.com/lukeramljak/charsibot/twitch)
[![Twitch Bot](https://github.com/lukeramljak/charsibot/actions/workflows/twitch.yml/badge.svg)](https://github.com/lukeramljak/charsibot/actions/workflows/twitch.yml)

A multi-service bot application for [Charsibel](https://twitch.tv/charsibel)'s Discord server and Twitch channel.

## Architecture

This project contains two services:

- **Discord Bot** (`discord/`) - Handles Discord server interactions and commands (Node.js/Bun)
- **Twitch Bot** (`twitch/`) - Manages Twitch chat commands, channel point redemptions, API server, and stream overlay (Go)

The web frontend (`twitch/web/`) is a SvelteKit SPA embedded directly in the Go binary at build time.

## Prerequisites

- Bun
- Go 1.25+
- Docker
- Task

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
   task dev
   ```

   This will start the Go backend and API server on port 8081.

4. **Twitch Web Frontend**:

   ```bash
   cd twitch/web
   pnpm install
   pnpm dev
   ```

   Access the frontend at `http://localhost:5173`. The Vite dev server proxies `/events` and `/api` to the Go server at `localhost:8081`.

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
