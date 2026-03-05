package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/robfig/cron/v3"
)

type Config struct {
	TGChatID     int64
	TGToken      string
	OfficeID     string
	CurrencyCode string
	CronSchedule string
}

func LoadConfig() (*Config, error) {
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

	currencyCode := os.Getenv("CURRENCY_CODE")
	if currencyCode == "" {
		return nil, fmt.Errorf("CURRENCY_CODE is not set")
	}

	cronSchedule := os.Getenv("CRON_SCHEDULE")
	if cronSchedule == "" {
		return nil, fmt.Errorf("CRON_SCHEDULE is not set")
	}
	if _, err := cron.ParseStandard(cronSchedule); err != nil {
		return nil, fmt.Errorf("CRON_SCHEDULE %q is not a valid cron string: %w", cronSchedule, err)
	}

	return &Config{
		TGChatID:     tgChatID,
		TGToken:      tgToken,
		OfficeID:     officeID,
		CurrencyCode: currencyCode,
		CronSchedule: cronSchedule,
	}, nil
}
