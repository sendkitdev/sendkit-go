package sendkit

import (
	"context"
	"encoding/json"
	"fmt"
)

// EmailsService handles communication with the email related methods of the SendKit API.
type EmailsService struct {
	client *Client
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
}

// SendEmailParams contains the parameters for sending a structured email.
type SendEmailParams struct {
	From        string            `json:"from"`
	To          []string          `json:"to"`
	Subject     string            `json:"subject"`
	HTML        string            `json:"html,omitempty"`
	Text        string            `json:"text,omitempty"`
	CC          []string          `json:"cc,omitempty"`
	BCC         []string          `json:"bcc,omitempty"`
	ReplyTo     string            `json:"reply_to,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	ScheduledAt string            `json:"scheduled_at,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
}

// SendEmailResponse is the response from sending an email.
type SendEmailResponse struct {
	ID string `json:"id"`
}

// SendMimeEmailParams contains the parameters for sending a raw MIME email.
type SendMimeEmailParams struct {
	EnvelopeFrom string `json:"envelope_from"`
	EnvelopeTo   string `json:"envelope_to"`
	RawMessage   string `json:"raw_message"`
}

// SendMimeEmailResponse is the response from sending a MIME email.
type SendMimeEmailResponse struct {
	ID string `json:"id"`
}

// Send sends a structured email via the SendKit API.
func (s *EmailsService) Send(ctx context.Context, params *SendEmailParams) (*SendEmailResponse, error) {
	body, err := s.client.doRequest(ctx, "POST", "/emails", params)
	if err != nil {
		return nil, err
	}

	var resp SendEmailResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("sendkit: failed to parse response: %w", err)
	}

	return &resp, nil
}

// SendMime sends a raw MIME email via the SendKit API.
func (s *EmailsService) SendMime(ctx context.Context, params *SendMimeEmailParams) (*SendMimeEmailResponse, error) {
	body, err := s.client.doRequest(ctx, "POST", "/emails/mime", params)
	if err != nil {
		return nil, err
	}

	var resp SendMimeEmailResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("sendkit: failed to parse response: %w", err)
	}

	return &resp, nil
}
