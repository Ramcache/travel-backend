package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Ramcache/travel-backend/internal/models"
	"io"
	"net/http"
	"net/url"
)

var (
	httpPostForm = http.PostForm
	httpPostJSON = func(url, contentType string, body io.Reader) (*http.Response, error) {
		return http.Post(url, contentType, body)
	}
)

type TelegramClient struct {
	Token  string
	ChatID string
}

func (t *TelegramClient) Create(ctx context.Context, hotel *models.Hotel) error {
	//TODO implement me
	panic("implement me")
}

func (t *TelegramClient) Get(ctx context.Context, id int) (*models.Hotel, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TelegramClient) List(ctx context.Context) ([]models.Hotel, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TelegramClient) Update(ctx context.Context, hotel *models.Hotel) error {
	//TODO implement me
	panic("implement me")
}

func (t *TelegramClient) Delete(ctx context.Context, id int) error {
	//TODO implement me
	panic("implement me")
}

func (t *TelegramClient) Attach(ctx context.Context, th *models.TripHotel) error {
	//TODO implement me
	panic("implement me")
}

func (t *TelegramClient) ListByTrip(ctx context.Context, tripID int) ([]models.Hotel, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TelegramClient) ClearByTrip(ctx context.Context, tripID int) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func NewTelegramClient(token, chatID string) *TelegramClient {
	return &TelegramClient{Token: token, ChatID: chatID}
}

func (t *TelegramClient) SendMessage(text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.Token)

	resp, err := httpPostForm(apiURL, url.Values{
		"chat_id":    {t.ChatID},
		"text":       {text},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram send failed: %s", resp.Status)
	}
	return nil
}

func (t *TelegramClient) SendMessageWithButton(text, buttonText, buttonURL string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.Token)

	payload := map[string]any{
		"chat_id":    t.ChatID,
		"text":       text,
		"parse_mode": "HTML",
		"reply_markup": map[string]any{
			"inline_keyboard": [][]map[string]any{
				{
					{
						"text": buttonText,
						"url":  buttonURL,
					},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := httpPostJSON(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram send failed: %s, body: %s", resp.Status, string(b))
	}
	return nil
}
