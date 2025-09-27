package helpers

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestNewTelegramClient(t *testing.T) {
	client := NewTelegramClient("token", "chat")
	if client.Token != "token" || client.ChatID != "chat" {
		t.Fatalf("unexpected client fields: %+v", client)
	}
}

func TestSendMessageSuccess(t *testing.T) {
	client := NewTelegramClient("token", "chat")

	called := false
	httpPostForm = func(url string, data url.Values) (*http.Response, error) {
		called = true
		if data.Get("chat_id") != "chat" {
			t.Fatalf("unexpected chat id %s", data.Get("chat_id"))
		}
		if data.Get("text") != "hello" {
			t.Fatalf("unexpected text %s", data.Get("text"))
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(""))}, nil
	}
	t.Cleanup(func() { httpPostForm = http.PostForm })

	if err := client.SendMessage("hello"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Fatal("expected httpPostForm to be called")
	}
}

func TestSendMessageHTTPError(t *testing.T) {
	client := NewTelegramClient("token", "chat")

	httpPostForm = func(string, url.Values) (*http.Response, error) {
		return nil, errors.New("network")
	}
	t.Cleanup(func() { httpPostForm = http.PostForm })

	if err := client.SendMessage("hello"); err == nil || !strings.Contains(err.Error(), "network") {
		t.Fatalf("expected network error, got %v", err)
	}
}

func TestSendMessageNon200(t *testing.T) {
	client := NewTelegramClient("token", "chat")

	httpPostForm = func(string, url.Values) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad Request", Body: io.NopCloser(strings.NewReader("bad"))}, nil
	}
	t.Cleanup(func() { httpPostForm = http.PostForm })

	if err := client.SendMessage("hello"); err == nil || !strings.Contains(err.Error(), "telegram send failed") {
		t.Fatalf("expected telegram send failure, got %v", err)
	}
}

func TestSendMessageWithButtonSuccess(t *testing.T) {
	client := NewTelegramClient("token", "chat")

	httpPostJSON = func(url, contentType string, body io.Reader) (*http.Response, error) {
		if !strings.Contains(url, "token") {
			t.Fatalf("unexpected url %s", url)
		}
		if contentType != "application/json" {
			t.Fatalf("unexpected content type %s", contentType)
		}
		data, err := io.ReadAll(body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if !strings.Contains(string(data), "button") {
			t.Fatalf("expected button text in payload, got %s", string(data))
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(""))}, nil
	}
	t.Cleanup(func() {
		httpPostJSON = func(url, contentType string, body io.Reader) (*http.Response, error) {
			return http.Post(url, contentType, body)
		}
	})

	if err := client.SendMessageWithButton("hi", "button", "https://example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendMessageWithButtonHTTPError(t *testing.T) {
	client := NewTelegramClient("token", "chat")

	httpPostJSON = func(string, string, io.Reader) (*http.Response, error) {
		return nil, errors.New("network")
	}
	t.Cleanup(func() {
		httpPostJSON = func(url, contentType string, body io.Reader) (*http.Response, error) {
			return http.Post(url, contentType, body)
		}
	})

	if err := client.SendMessageWithButton("hi", "button", "https://example.com"); err == nil || !strings.Contains(err.Error(), "network") {
		t.Fatalf("expected network error, got %v", err)
	}
}

func TestSendMessageWithButtonNon200(t *testing.T) {
	client := NewTelegramClient("token", "chat")

	httpPostJSON = func(string, string, io.Reader) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad Request", Body: io.NopCloser(strings.NewReader("bad"))}, nil
	}
	t.Cleanup(func() {
		httpPostJSON = func(url, contentType string, body io.Reader) (*http.Response, error) {
			return http.Post(url, contentType, body)
		}
	})

	if err := client.SendMessageWithButton("hi", "button", "https://example.com"); err == nil || !strings.Contains(err.Error(), "telegram send failed") {
		t.Fatalf("expected telegram send failure, got %v", err)
	}
}
