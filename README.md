
# Sber~~RateSell~~SellRateNotifier

A small tool that notifies you of the exchange rate at your local Sberbank office.

## How It Works

At the scheduled time, the service fetches the current exchange rate for a given Sberbank office via the public Sberbank API and sends the sell rate to a configured Telegram chat.

## Requirements

- Docker (or Go 1.25+)
- A Telegram bot token ([create one via @BotFather](https://t.me/BotFather))
- Your Telegram chat ID
- Your Sberbank office ID

## Finding Your Office ID

The office ID can be found at https://www.sberbank.ru/ru/oib?tab=vsp. Zoom at your branch and look into page sources.
