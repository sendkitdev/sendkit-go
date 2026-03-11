package sendkit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewClient_WithAPIKey(t *testing.T) {
	client, err := NewClient("sk_test_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client to not be nil")
	}
}

func TestNewClient_MissingAPIKey(t *testing.T) {
	os.Unsetenv("SENDKIT_API_KEY")

	_, err := NewClient("")
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestNewClient_FromEnvVariable(t *testing.T) {
	os.Setenv("SENDKIT_API_KEY", "sk_from_env")
	defer os.Unsetenv("SENDKIT_API_KEY")

	client, err := NewClient("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client to not be nil")
	}
}

func TestNewClient_CustomBaseURL(t *testing.T) {
	client, err := NewClient("sk_test_123", WithBaseURL("https://custom.api.com"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.baseURL != "https://custom.api.com" {
		t.Fatalf("expected base URL to be https://custom.api.com, got %s", client.baseURL)
	}
}

func TestEmails_Send(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails" {
			t.Errorf("expected path /emails, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer sk_test_123" {
			t.Errorf("expected Authorization header to be Bearer sk_test_123, got %s", r.Header.Get("Authorization"))
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		if payload["from"] != "sender@example.com" {
			t.Errorf("expected from to be sender@example.com, got %v", payload["from"])
		}
		if payload["subject"] != "Test Email" {
			t.Errorf("expected subject to be Test Email, got %v", payload["subject"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "email-uuid-123"})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	resp, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Email",
		HTML:    "<p>Hello</p>",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "email-uuid-123" {
		t.Fatalf("expected ID email-uuid-123, got %s", resp.ID)
	}
}

func TestEmails_Send_SnakeCaseFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		if payload["reply_to"] != "reply@example.com" {
			t.Errorf("expected reply_to to be reply@example.com, got %v", payload["reply_to"])
		}
		if payload["scheduled_at"] != "2026-03-01T10:00:00Z" {
			t.Errorf("expected scheduled_at to be 2026-03-01T10:00:00Z, got %v", payload["scheduled_at"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "email-uuid-456"})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	_, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:        "sender@example.com",
		To:          []string{"recipient@example.com"},
		Subject:     "Test",
		HTML:        "<p>Hi</p>",
		ReplyTo:     "reply@example.com",
		ScheduledAt: "2026-03-01T10:00:00Z",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmails_SendMime(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails/mime" {
			t.Errorf("expected path /emails/mime, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		if payload["envelope_from"] != "sender@example.com" {
			t.Errorf("expected envelope_from to be sender@example.com, got %v", payload["envelope_from"])
		}
		if payload["envelope_to"] != "recipient@example.com" {
			t.Errorf("expected envelope_to to be recipient@example.com, got %v", payload["envelope_to"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "mime-uuid-789"})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	resp, err := client.Emails.SendMime(context.Background(), &SendMimeEmailParams{
		EnvelopeFrom: "sender@example.com",
		EnvelopeTo:   "recipient@example.com",
		RawMessage:   "From: sender@example.com\r\nTo: recipient@example.com\r\n\r\nHello",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "mime-uuid-789" {
		t.Fatalf("expected ID mime-uuid-789, got %s", resp.ID)
	}
}

func TestEmails_Send_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]any{
			"name":       "validation_error",
			"message":    "The to field is required.",
			"statusCode": 422,
		})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	_, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:    "sender@example.com",
		To:      []string{},
		Subject: "Test",
		HTML:    "<p>Hi</p>",
	})

	if err == nil {
		t.Fatal("expected error for API failure")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Name != "validation_error" {
		t.Errorf("expected name validation_error, got %s", apiErr.Name)
	}
	if apiErr.StatusCode != 422 {
		t.Errorf("expected status code 422, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "The to field is required." {
		t.Errorf("expected message 'The to field is required.', got %s", apiErr.Message)
	}
}

func TestNewSendEmailParams(t *testing.T) {
	params := NewSendEmailParams("sender@example.com", "recipient@example.com", "Hello")

	if params.From != "sender@example.com" {
		t.Errorf("expected from sender@example.com, got %s", params.From)
	}
	if len(params.To) != 1 || params.To[0] != "recipient@example.com" {
		t.Errorf("expected to [recipient@example.com], got %v", params.To)
	}
	if params.Subject != "Hello" {
		t.Errorf("expected subject Hello, got %s", params.Subject)
	}
}

func TestNewSendEmailParams_DisplayName(t *testing.T) {
	params := NewSendEmailParams("Support <sender@example.com>", "Bob <recipient@example.com>", "Hello")

	if params.From != "Support <sender@example.com>" {
		t.Errorf("expected from with display name, got %s", params.From)
	}
	if params.To[0] != "Bob <recipient@example.com>" {
		t.Errorf("expected to with display name, got %s", params.To[0])
	}
}

func TestEmails_Send_CustomBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "custom-url-123"})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	resp, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		HTML:    "<p>Hi</p>",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "custom-url-123" {
		t.Fatalf("expected ID custom-url-123, got %s", resp.ID)
	}
}
