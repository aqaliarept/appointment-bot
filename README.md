# Appointment Availability Bot

A Telegram bot that monitors appointment availability and notifies users when slots become available.

## Features

- Real-time monitoring of appointment availability
- Automatic notifications when availability changes
- Optional status updates every 30 minutes
- Support for multiple users
- Manual availability checks
- Detailed logging

## Setup

1. Clone the repository:

```bash
git clone git@github.com:aqaliarept/appointment-bot.git
cd appointment-bot
```

2. Create a `.env` file in the project root with your Telegram bot token:

```bash
TELEGRAM_BOT_TOKEN=your_bot_token_here
```

3. Run the bot:

```bash
# Run in bot mode (continuous monitoring)
go run main.go -bot

# Run once to check current availability
go run main.go
```

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

## License

MIT License
