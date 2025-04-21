# charsibot

## Overview

charsibot is a Discord bot used in [Charsibel](https://twitch.tv/charsibel)'s server. This is not a general purpose bot, and is intended to be used in a specific server.

## Prerequisites

- Go 1.21+
- Docker

## Installation

1. Clone the repository: `git clone https://github.com/lukeramljak/charsibot.git`
2. Install dependencies: `go mod tidy`
3. Create an `.env` file using the provided `.env.example`
4. During development, you can use `go run main.go`
5. To build the bot: `go build -o charsibot`

## Deployment & Updates

1. SSH into the server
2. Navigate to the bot directory: `cd charsibot`
3. Pull the latest changes: `git pull`
4. Rebuild and restart the Docker image: `docker compose up --build -d`
