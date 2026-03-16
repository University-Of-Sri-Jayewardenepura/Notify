package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultBaseURL = "https://discord.com/api/webhooks"

type Client struct {
	baseURL      string
	webhookID    string
	webhookToken string
	httpClient   *http.Client
}

func NewClient(baseURL string, webhookID string, webhookToken string, httpClient *http.Client) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:      strings.TrimRight(baseURL, "/"),
		webhookID:    webhookID,
		webhookToken: webhookToken,
		httpClient:   httpClient,
	}
}

func (c *Client) Send(ctx context.Context, payload WebhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal discord payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/"+c.webhookID+"/"+c.webhookToken, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create discord request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send discord request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	return nil
}
