package sberparser

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

const rateURL = "https://www.sberbank.ru/proxy/services/rates/public/v4/branchesRates"

var parsedRateURL *url.URL

func init() {
	var err error
	parsedRateURL, err = url.Parse(rateURL)
	if err != nil {
		panic(fmt.Sprintf("invalid rateURL constant: %v", err))
	}
}

type RateEntry struct {
	RangeAmountBottom float64 `json:"rangeAmountBottom"`
	RangeAmountUpper  float64 `json:"rangeAmountUpper"`
	RateSell          float64 `json:"rateSell"`
	RateBuy           float64 `json:"rateBuy"`
}

type CurrencyRate struct {
	StartDateTime int64       `json:"startDateTime"`
	LotSize       int         `json:"lotSize"`
	RateList      []RateEntry `json:"rateList"`
}

type BranchRates struct {
	ID    int64                   `json:"id"`
	Rates map[string]CurrencyRate `json:"rates"`
}

func FetchLocalRates(ctx context.Context, client *http.Client, officeIDs []string, currencyCode string) ([]BranchRates, error) {
	u := *parsedRateURL

	q := u.Query()
	q.Set("rateType", "ERNP-1")
	q.Set("isoCode", currencyCode)
	for _, id := range officeIDs {
		q.Add("id[]", id)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", "https://www.sberbank.ru/")

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	slog.Info("got responce from sberbank API")

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", response.Status)
	}

	var result []BranchRates
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
