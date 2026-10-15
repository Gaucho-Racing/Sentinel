package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrNotConfigured = errors.New("notify is not configured")

var httpClient = &http.Client{Timeout: 10 * time.Second}

type Client struct {
	BaseURL string
	Token   string
}

type Email struct {
	EntityID string
	SenderID string
	Category string
	Subject  string
	Body     string
	HTML     string
}

type sendRequest struct {
	Channel    string       `json:"channel"`
	Category   string       `json:"category,omitempty"`
	Recipients recipients   `json:"recipients"`
	Email      emailContent `json:"email"`
}

type recipients struct {
	EntityIDs []string `json:"entity_ids"`
	GroupIDs  []string `json:"group_ids"`
}

type emailContent struct {
	SenderID string `json:"sender_id,omitempty"`
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	HTML     string `json:"html,omitempty"`
}

func (c Client) SendEmail(ctx context.Context, email Email) error {
	if c.BaseURL == "" || c.Token == "" {
		return ErrNotConfigured
	}
	payload, err := json.Marshal(sendRequest{
		Channel:    "EMAIL",
		Category:   email.Category,
		Recipients: recipients{EntityIDs: []string{email.EntityID}, GroupIDs: []string{}},
		Email: emailContent{
			SenderID: email.SenderID,
			Subject:  email.Subject,
			Body:     email.Body,
			HTML:     email.HTML,
		},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/messages", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send to notify: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var notifyErr struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &notifyErr) == nil && notifyErr.Error != "" {
		return fmt.Errorf("notify rejected the message: [%d] %s", resp.StatusCode, notifyErr.Error)
	}
	return fmt.Errorf("notify rejected the message: [%d] %s", resp.StatusCode, strings.TrimSpace(string(body)))
}
