package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	ErrFetchFailed  = errors.New("не удалось получить курсы валют")
	ErrDecodeFailed = errors.New("ошибка обработки ответа ЦБ РФ")
)

type CurrencyRate struct {
	USD float64 `json:"usd"`
	SAR float64 `json:"sar"` // риал
}

type CurrencyService struct {
	mu    sync.RWMutex
	rates CurrencyRate
	last  time.Time
	ttl   time.Duration
	log   *zap.SugaredLogger
	http  *http.Client
}

func NewCurrencyService(ttl time.Duration, log *zap.SugaredLogger) *CurrencyService {
	return &CurrencyService{
		ttl:  ttl,
		log:  log,
		http: &http.Client{Timeout: 5 * time.Second}, // безопасный таймаут
	}
}

func (s *CurrencyService) GetRates(ctx context.Context) (CurrencyRate, error) {
	// читаем кеш
	s.mu.RLock()
	if time.Since(s.last) < s.ttl {
		defer s.mu.RUnlock()
		return s.rates, nil
	}
	s.mu.RUnlock()

	// новый запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.cbr-xml-daily.ru/daily_json.js", nil)
	if err != nil {
		s.log.Errorw("currency_request_build_failed", "err", err)
		return CurrencyRate{}, ErrFetchFailed
	}

	resp, err := s.http.Do(req)
	if err != nil {
		s.log.Errorw("currency_fetch_failed", "err", err)
		return CurrencyRate{}, ErrFetchFailed
	}
	defer resp.Body.Close()

	var data struct {
		Valute map[string]struct {
			Value float64 `json:"Value"`
		} `json:"Valute"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		s.log.Errorw("currency_decode_failed", "err", err)
		return CurrencyRate{}, ErrDecodeFailed
	}

	rates := CurrencyRate{
		USD: data.Valute["USD"].Value,
		SAR: data.Valute["SAR"].Value,
	}

	// обновляем кеш
	s.mu.Lock()
	s.rates = rates
	s.last = time.Now()
	s.mu.Unlock()

	s.log.Infow("currency_rates_updated", "usd", rates.USD, "sar", rates.SAR)

	return rates, nil
}
