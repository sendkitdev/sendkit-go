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

func TestEmails_Send_MultipleRecipients(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		to, ok := payload["to"].([]any)
		if !ok {
			t.Fatal("expected to to be an array")
		}
		if len(to) != 3 {
			t.Fatalf("expected 3 recipients, got %d", len(to))
		}
		if to[0] != "alice@example.com" {
			t.Errorf("expected first recipient alice@example.com, got %v", to[0])
		}
		if to[1] != "bob@example.com" {
			t.Errorf("expected second recipient bob@example.com, got %v", to[1])
		}
		if to[2] != "carol@example.com" {
			t.Errorf("expected third recipient carol@example.com, got %v", to[2])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "multi-recipient-123"})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	resp, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:    "sender@example.com",
		To:      []string{"alice@example.com", "bob@example.com", "carol@example.com"},
		Subject: "Test Multiple Recipients",
		HTML:    "<p>Hello everyone</p>",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "multi-recipient-123" {
		t.Fatalf("expected ID multi-recipient-123, got %s", resp.ID)
	}
}

func TestEmails_Send_CcAndBcc(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		cc, ok := payload["cc"].([]any)
		if !ok {
			t.Fatal("expected cc to be an array")
		}
		if len(cc) != 2 {
			t.Fatalf("expected 2 cc recipients, got %d", len(cc))
		}
		if cc[0] != "cc1@example.com" {
			t.Errorf("expected first cc cc1@example.com, got %v", cc[0])
		}
		if cc[1] != "cc2@example.com" {
			t.Errorf("expected second cc cc2@example.com, got %v", cc[1])
		}

		bcc, ok := payload["bcc"].([]any)
		if !ok {
			t.Fatal("expected bcc to be an array")
		}
		if len(bcc) != 1 {
			t.Fatalf("expected 1 bcc recipient, got %d", len(bcc))
		}
		if bcc[0] != "bcc@example.com" {
			t.Errorf("expected bcc bcc@example.com, got %v", bcc[0])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "cc-bcc-123"})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	resp, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test CC and BCC",
		HTML:    "<p>Hello</p>",
		CC:      []string{"cc1@example.com", "cc2@example.com"},
		BCC:     []string{"bcc@example.com"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "cc-bcc-123" {
		t.Fatalf("expected ID cc-bcc-123, got %s", resp.ID)
	}
}

func TestEmails_Send_Attachments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		attachments, ok := payload["attachments"].([]any)
		if !ok {
			t.Fatal("expected attachments to be an array")
		}
		if len(attachments) != 2 {
			t.Fatalf("expected 2 attachments, got %d", len(attachments))
		}

		att1, ok := attachments[0].(map[string]any)
		if !ok {
			t.Fatal("expected attachment to be an object")
		}
		if att1["filename"] != "report.pdf" {
			t.Errorf("expected filename report.pdf, got %v", att1["filename"])
		}
		if att1["content"] != "base64-pdf-content" {
			t.Errorf("expected content base64-pdf-content, got %v", att1["content"])
		}
		if att1["content_type"] != "application/pdf" {
			t.Errorf("expected content_type application/pdf, got %v", att1["content_type"])
		}

		att2, ok := attachments[1].(map[string]any)
		if !ok {
			t.Fatal("expected attachment to be an object")
		}
		if att2["filename"] != "image.png" {
			t.Errorf("expected filename image.png, got %v", att2["filename"])
		}
		if att2["content"] != "base64-png-content" {
			t.Errorf("expected content base64-png-content, got %v", att2["content"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "attachment-123"})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	resp, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Attachments",
		HTML:    "<p>See attached</p>",
		Attachments: []Attachment{
			{Filename: "report.pdf", Content: "base64-pdf-content", ContentType: "application/pdf"},
			{Filename: "image.png", Content: "base64-png-content"},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "attachment-123" {
		t.Fatalf("expected ID attachment-123, got %s", resp.ID)
	}
}

func TestEmails_Send_Tags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		tags, ok := payload["tags"].([]any)
		if !ok {
			t.Fatal("expected tags to be an array")
		}
		if len(tags) != 2 {
			t.Fatalf("expected 2 tags, got %d", len(tags))
		}

		tag1, ok := tags[0].(map[string]any)
		if !ok {
			t.Fatal("expected tag to be an object")
		}
		if tag1["name"] != "category" {
			t.Errorf("expected first tag name category, got %v", tag1["name"])
		}
		if tag1["value"] != "welcome" {
			t.Errorf("expected first tag value welcome, got %v", tag1["value"])
		}

		tag2, ok := tags[1].(map[string]any)
		if !ok {
			t.Fatal("expected tag to be an object")
		}
		if tag2["name"] != "campaign" {
			t.Errorf("expected second tag name campaign, got %v", tag2["name"])
		}
		if tag2["value"] != "onboarding" {
			t.Errorf("expected second tag value onboarding, got %v", tag2["value"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "tags-123"})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	resp, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Tags",
		HTML:    "<p>Hello</p>",
		Tags: []Tag{
			{Name: "category", Value: "welcome"},
			{Name: "campaign", Value: "onboarding"},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "tags-123" {
		t.Fatalf("expected ID tags-123, got %s", resp.ID)
	}
}

func TestEmails_Send_TextField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		if payload["text"] != "Hello plain text" {
			t.Errorf("expected text to be Hello plain text, got %v", payload["text"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "text-123"})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	resp, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Text Field",
		Text:    "Hello plain text",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "text-123" {
		t.Fatalf("expected ID text-123, got %s", resp.ID)
	}
}

func TestEmails_Send_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		headers, ok := payload["headers"].(map[string]any)
		if !ok {
			t.Fatal("expected headers to be an object")
		}
		if headers["X-Custom-Header"] != "value" {
			t.Errorf("expected X-Custom-Header to be value, got %v", headers["X-Custom-Header"])
		}
		if headers["X-Track"] != "123" {
			t.Errorf("expected X-Track to be 123, got %v", headers["X-Track"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "headers-123"})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	resp, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Headers",
		HTML:    "<p>Hello</p>",
		Headers: map[string]string{"X-Custom-Header": "value", "X-Track": "123"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "headers-123" {
		t.Fatalf("expected ID headers-123, got %s", resp.ID)
	}
}

func TestEmails_Send_OmitsNullFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)

		omittedKeys := []string{"text", "cc", "bcc", "reply_to", "headers", "tags", "scheduled_at", "attachments"}
		for _, key := range omittedKeys {
			if _, ok := payload[key]; ok {
				t.Errorf("expected key %q to be omitted from JSON body, but it was present", key)
			}
		}

		if payload["from"] != "sender@example.com" {
			t.Errorf("expected from to be sender@example.com, got %v", payload["from"])
		}
		if payload["subject"] != "Required Only" {
			t.Errorf("expected subject to be Required Only, got %v", payload["subject"])
		}
		if payload["html"] != "<p>Hello</p>" {
			t.Errorf("expected html to be <p>Hello</p>, got %v", payload["html"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "omit-123"})
	}))
	defer server.Close()

	client, _ := NewClient("sk_test_123", WithBaseURL(server.URL))
	resp, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Required Only",
		HTML:    "<p>Hello</p>",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "omit-123" {
		t.Fatalf("expected ID omit-123, got %s", resp.ID)
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
