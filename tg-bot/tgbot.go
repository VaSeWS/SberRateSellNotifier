package tgbot

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotWrapper struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func NewBotWrapper(tgApiToken string, chatID int64) (*BotWrapper, error) {
	bot, err := tgbotapi.NewBotAPI(tgApiToken)
	if err != nil {
		return nil, err
	}
	return &BotWrapper{bot, chatID}, nil
}

func (b *BotWrapper) SendMessage(msg string) error {
	_, err := b.bot.Send(tgbotapi.NewMessage(b.chatID, msg))
	if err != nil {
		return err
	}
	slog.Info("successfully sent message", "chatID", b.chatID, "message", msg)
	return nil
}
