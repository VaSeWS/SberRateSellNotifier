package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"

	sberparser "github.com/VaSeWS/SberRateSellNotifier/sber-parser"
	tgbot "github.com/VaSeWS/SberRateSellNotifier/tg-bot"
)

const currencyCode = "USD"

type Config struct {
	TGChatID int64
	TGToken  string
	OfficeID string
}

func loadConfig() (*Config, error) {
	tgChatID, err := strconv.ParseInt(os.Getenv("TG_CHAT_ID"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TG_CHAT_ID: %w", err)
	}
	tgToken := os.Getenv("TELEGRAM_API_TOKEN")
	if tgToken == "" {
		return nil, fmt.Errorf("TELEGRAM_API_TOKEN is not set")
	}
	officeID := os.Getenv("OFFICE_ID")
	if officeID == "" {
		return nil, fmt.Errorf("OFFICE_ID is not set")
	}
	return &Config{
		TGChatID: tgChatID,
		TGToken:  tgToken,
		OfficeID: officeID,
	}, nil
}

func main() {
	conf, err := loadConfig()
	if err != nil {
		slog.Error("failed to load a config", "error", err)
		os.Exit(1)
	}

	bot, err := tgbot.NewBotWrapper(conf.TGToken, conf.TGChatID)
	if err != nil {
		slog.Error("failed to initialize the bot", "error", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	offices := []string{conf.OfficeID}
	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		cancel() // signal everything to stop
	}()

	cr := cron.New()
	_, err = cr.AddFunc("0 12 * * *", func() {
		resp, err := sberparser.FetchLocalRates(ctx, client, offices, "USD")
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			msg := fmt.Sprintln("error fetching rate: ", err)
			bot.SendMessage(msg)
			slog.Error("error fetching rate", "error", err)
			return
		}
		for _, rate := range resp {

			rateData, ok := rate.Rates[currencyCode]
			if !ok || len(rateData.RateList) == 0 {
				slog.Error("no rates for office", slog.Int64("officeID", rate.ID))
				continue
			}

			sellRate := rateData.RateList[0].RateSell
			msg := fmt.Sprintf("Office ID: %d, Sell rate: %.2f", rate.ID, sellRate)
			if err := bot.SendMessage(msg); err != nil {
				slog.Error("failed to send message", "error", err)
			}

		}
	})
	if err != nil {
		slog.Error("invalid cron spec", "error", err)
		os.Exit(1)
	}
	cr.Start()
	defer cr.Stop()

	<-ctx.Done()
	slog.Info("Shutting down...")
}
