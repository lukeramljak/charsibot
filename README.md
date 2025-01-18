# charsibot

## Overview

charsibot is a Discord bot used in [Charsibel](https://twitch.tv/charsibel)'s server. This is not a general purpose bot, and is intended to be used in a specific server.

## Installation

1. Clone the repository: `git clone https://github.com/lukeramljak/charsibot.git`
2. Install Go if not already installed: [Go Installation Guide](https://golang.org/doc/install)
3. Install dependencies: `go mod tidy`
4. Create an `.env` file using the provided `.env.example`
5. During development, you can use `go run main.go`
6. To build the bot: `go build -o charsibot`

## Running on a Server

1. Build the binary: `go build -o charsibot`
2. Create a systemd service file `/etc/systemd/system/charsibot.service`:

```
[Unit]
Description=charsibot
After=network.target

[Service]
Type=simple
User=your_user
WorkingDirectory=/path/to/bot
ExecStart=/path/to/bot/charsibot
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

3. Enable and start the service:
   - `sudo systemctl enable charsibot`
   - `sudo systemctl start charsibot`
4. View logs: `sudo journalctl -u charsibot`
5. Manage the service:
   - Stop: `sudo systemctl stop charsibot`
   - Restart: `sudo systemctl restart charsibot`
   - Check status: `sudo systemctl status charsibot`

## Deployment & Updates

When pushing updates to production:

1. SSH into the server
2. Navigate to the bot directory: `cd /path/to/bot`
3. Pull the latest changes: `git pull`
4. Rebuild the binary: `go build -o charsibot`
5. Restart the service: `sudo systemctl restart charsibot`
6. Verify the bot is running: `sudo systemctl status charsibot`

To view recent logs after deployment:

```bash
sudo journalctl -u charsibot -n 50 --no-pager
```
