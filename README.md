# Appointment Availability Bot

A Telegram bot that monitors appointment availability and notifies users when slots become available.

## Features

- Real-time monitoring of appointment availability
- Automatic notifications when availability changes
- Optional status updates every 30 minutes
- Support for multiple users
- Manual availability checks
- Detailed logging
- Configurable check interval

## Setup

### Running with Docker

1. Pull the latest image:

```bash
docker pull ghcr.io/aqaliarept/appointment-bot:latest
```

2. Run the bot with your Telegram token (in detached mode):

```bash
docker run -d --name appointment-bot -e TELEGRAM_BOT_TOKEN=your_bot_token_here ghcr.io/aqaliarept/appointment-bot:latest
```

You can configure the check interval using the `CHECK_INTERVAL` environment variable:

```bash
# Check every 5 minutes
docker run -d --name appointment-bot \
  -e TELEGRAM_BOT_TOKEN=your_bot_token_here \
  -e CHECK_INTERVAL=5m \
  ghcr.io/aqaliarept/appointment-bot:latest

# Check every 30 minutes
docker run -d --name appointment-bot \
  -e TELEGRAM_BOT_TOKEN=your_bot_token_here \
  -e CHECK_INTERVAL=30m \
  ghcr.io/aqaliarept/appointment-bot:latest
```

### Managing the Docker container

```bash
# View logs
docker logs appointment-bot
# Follow logs in real-time
docker logs -f appointment-bot

# Stop the bot
docker stop appointment-bot

# Start the bot again
docker start appointment-bot

# Remove the container
docker rm appointment-bot

# View container status
docker ps -a | grep appointment-bot
```

### Running locally

1. Clone the repository:

```bash
git clone git@github.com:aqaliarept/appointment-bot.git
cd appointment-bot
```

2. Create a `.env` file in the project root with your Telegram bot token:

```bash
TELEGRAM_BOT_TOKEN=your_bot_token_here
CHECK_INTERVAL=10m  # Optional, defaults to 10 minutes
```

3. Run the bot:

```bash
# Run in bot mode (continuous monitoring)
go run main.go -bot

# Run once to check current availability
go run main.go
```

## Environment Variables

| Variable             | Description                          | Default  | Example Values                              |
| -------------------- | ------------------------------------ | -------- | ------------------------------------------- |
| `TELEGRAM_BOT_TOKEN` | Your Telegram bot token              | Required | `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11` |
| `CHECK_INTERVAL`     | Interval between availability checks | `10m`    | `30s`, `5m`, `1h`                           |

## Bot Commands

- 🔍 Check Availability - Check current appointment availability
- 📊 Status - Show your notification settings
- ⏰ Enable Status Updates - Get status updates every 30 minutes
- ⏳ Disable Status Updates - Only get notifications when availability changes

## How it Works

The bot continuously monitors appointment availability by checking two different appointment types. Users are automatically notified when:

1. Appointments become available
2. Appointments are no longer available
3. Every 30 minutes if status updates are enabled

## Development

The bot is written in Go and uses:

- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) for Telegram integration
- [godotenv](https://github.com/joho/godotenv) for environment variable management

### Building Docker image locally

```bash
# Build the image
docker build -t appointment-bot .

# Run in detached mode
docker run -d --name appointment-bot \
  -e TELEGRAM_BOT_TOKEN=your_bot_token_here \
  -e CHECK_INTERVAL=5m \
  appointment-bot
```

### Testing the API

The bot checks two different appointment types using the Microsoft Bookings API. You can test these endpoints using the provided script:

```bash
# Make the script executable
chmod +x test-api.sh

# Run the tests
./test-api.sh
```

The script will make requests to both appointment types and display the responses. Make sure you have `curl` and `jq` installed:

```bash
# macOS (using Homebrew)
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# CentOS/RHEL
sudo yum install jq
```

Example API Response:

```json
{
  "staffAvailabilityResponse": [
    {
      "staffId": "...",
      "availabilityItems": [
        {
          "status": "BOOKINGSAVAILABILITYSTATUS_AVAILABLE",
          "startDateTime": {
            "dateTime": "2024-01-15T09:00:00.0000000",
            "timeZone": "FLE Standard Time"
          },
          "endDateTime": {
            "dateTime": "2024-01-15T09:30:00.0000000",
            "timeZone": "FLE Standard Time"
          },
          "availableCount": 1
        }
      ]
    }
  ]
}
```

## License

MIT License
