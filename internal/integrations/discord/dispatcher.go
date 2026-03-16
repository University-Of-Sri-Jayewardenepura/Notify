package discord

import (
	"context"
	"strings"
)

type Dispatcher struct {
	client *Client
}

func NewDispatcher(client *Client) *Dispatcher {
	return &Dispatcher{client: client}
}

func (d *Dispatcher) DispatchGitHubEvent(eventType string, payload map[string]any) error {
	webhookPayload, err := RenderGitHubEvent(eventType, payload)
	if err != nil {
		if shouldIgnoreRenderError(err) {
			return nil
		}
		return err
	}

	if d.client == nil {
		return nil
	}

	return d.client.Send(context.Background(), webhookPayload)
}

func shouldIgnoreRenderError(err error) bool {
	message := err.Error()
	return strings.HasPrefix(message, "skipping ") || strings.HasPrefix(message, "unhandled event type:")
}
