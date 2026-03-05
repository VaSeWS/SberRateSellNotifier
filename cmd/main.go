package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/VaSeWS/SberRateSellNotifier/internal/config"
	"github.com/VaSeWS/SberRateSellNotifier/internal/sberparser"
	"github.com/VaSeWS/SberRateSellNotifier/internal/tgbot"
)

func main() {
	conf, err := config.LoadConfig()
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
		cancel()
	}()

	cr := cron.New()
	_, err = cr.AddFunc(conf.CronSchedule, func() {
		resp, err := sberparser.FetchLocalRates(ctx, client, offices, conf.CurrencyCode)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			msg := fmt.Sprintf("error fetching rate: %v", err)
			if sendErr := bot.SendMessage(msg); sendErr != nil {
				slog.Error("error sending error message", "error", sendErr)
			}
			slog.Error("error fetching rate", "error", err)
			return
		}
		for _, rate := range resp {
			rateData, ok := rate.Rates[conf.CurrencyCode]
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

	<-ctx.Done()
	stopCtx := cr.Stop()
	<-stopCtx.Done()
	slog.Info("Shutting down...")
}
