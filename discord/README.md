# charsibot

## Overview

charsibot is a Discord bot used in [Charsibel](https://twitch.tv/charsibel)'s server. This is not a general purpose bot, and is intended to be used in a specific server.

## Prerequisites

- Go 1.21+
- Docker

## Installation

**Note: This service is now part of a multi-service setup. See the [root README](../README.md) for complete setup instructions.**

1. Clone the repository: `git clone https://github.com/lukeramljak/charsibot.git`
2. Install dependencies: `go work sync` (from root directory)
3. Set up environment: `cp .env.example .env` (from root directory)
4. During development, you can use `go run ./cmd/bot` (from this directory)
5. To build the bot: `go build -o charsibot ./cmd/bot`

## Deployment & Updates

**Note: This service is now part of a multi-service setup. See the [root README](../README.md) for full deployment instructions.**

### Quick Deployment:
1. SSH into the server
2. Navigate to the root directory: `cd charsibot`
3. Pull the latest changes: `git pull`
4. Deploy all services: `./deploy.sh` or just Discord: `./deploy.sh discord`
