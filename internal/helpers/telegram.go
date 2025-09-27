package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
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
