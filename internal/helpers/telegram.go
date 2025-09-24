package helpers

import (
	"fmt"
	"net/http"
	"net/url"
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

	resp, err := http.PostForm(apiURL, url.Values{
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
