# charsibot

[![Go Version](https://img.shields.io/github/go-mod/go-version/lukeramljak/charsibot?filename=go.mod)](https://github.com/lukeramljak/charsibot/blob/main/go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/lukeramljak/charsibot)](https://goreportcard.com/report/github.com/lukeramljak/charsibot)
[![Twitch Bot](https://github.com/lukeramljak/charsibot/actions/workflows/ci.yml/badge.svg)](https://github.com/lukeramljak/charsibot/actions/workflows/ci.yml)

Twitch bot and overlay for [Charsibel](https://twitch.tv/charsibel)

## Prerequisites

- Go 1.25+
- pnpm
- Docker
- Task

## Development Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/lukeramljak/charsibot.git
   cd charsibot
   ```

2. **Twitch Bot**:

   ```bash
   go mod download
   task dev
   ```

   This will start the Go backend and API server on port 8081.

3. **Twitch Web Frontend**:

   ```bash
   cd web
   pnpm install
   pnpm dev
   ```

   Access the frontend at `http://localhost:5173`. The Vite dev server proxies `/events` and `/api` to the Go server at `localhost:8081`.

## Environment Variables

   ```bash
   cp .env.example .env
   ```

## Database

The Twitch bot uses a SQLite database (`charsibot.db`) which will be created automatically in the project root.
