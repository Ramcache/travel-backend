package services

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
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
}

func NewCurrencyService(ttl time.Duration) *CurrencyService {
	return &CurrencyService{ttl: ttl}
}

func (s *CurrencyService) GetRates() (CurrencyRate, error) {
	s.mu.RLock()
	if time.Since(s.last) < s.ttl {
		defer s.mu.RUnlock()
		return s.rates, nil
	}
	s.mu.RUnlock()

	resp, err := http.Get("https://www.cbr-xml-daily.ru/daily_json.js")
	if err != nil {
		return CurrencyRate{}, err
	}
	defer resp.Body.Close()

	var data struct {
		Valute map[string]struct {
			Value float64 `json:"Value"`
		} `json:"Valute"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return CurrencyRate{}, err
	}

	rates := CurrencyRate{
		USD: data.Valute["USD"].Value,
		SAR: data.Valute["SAR"].Value,
	}

	s.mu.Lock()
	s.rates = rates
	s.last = time.Now()
	s.mu.Unlock()

	return rates, nil
}
