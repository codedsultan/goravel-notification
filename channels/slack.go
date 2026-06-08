package channels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	contractsnotification "github.com/codedsultan/goravel-notification/contracts"

	"github.com/goravel/framework/contracts/log"
)

// SlackChannel delivers notifications to Slack via incoming webhooks.
// Set the webhook URL by implementing RouteNotificationFor("slack") on your
// Notifiable model — typically stored as a per-user or per-workspace setting.
type SlackChannel struct {
	client *http.Client
	log    log.Log
}

// NewSlackChannel creates a SlackChannel with a sensible HTTP timeout.
func NewSlackChannel(logger log.Log) *SlackChannel {
	return &SlackChannel{
		client: &http.Client{Timeout: 10 * time.Second},
		log:    logger,
	}
}

// Name satisfies contracts.Channel.
func (c *SlackChannel) Name() string { return "slack" }

// Send satisfies contracts.Channel.
// It posts the SlackMessage JSON payload to the webhook URL returned by
// RouteNotificationFor("slack").
func (c *SlackChannel) Send(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	webhookURL := notifiable.RouteNotificationFor("slack")
	if webhookURL == "" {
		return fmt.Errorf("slack channel: %T.RouteNotificationFor(\"slack\") returned empty webhook URL", notifiable)
	}

	// Resolve the message payload.
	var msg contractsnotification.SlackMessage
	if sn, ok := n.(contractsnotification.SlackNotification); ok {
		msg = sn.ToSlack(notifiable)
	} else {
		msg = c.defaultMessage(n)
	}

	// Slack's webhook API expects the JSON shape below.
	body, err := json.Marshal(slackPayload(msg))
	if err != nil {
		return fmt.Errorf("slack channel: failed to marshal payload for %T: %w", n, err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack channel: failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack channel: HTTP request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack channel: webhook returned non-2xx status %d for %T", resp.StatusCode, n)
	}

	c.log.Debugf("notifications: delivered %T to Slack (status=%d)", n, resp.StatusCode)
	return nil
}

// defaultMessage builds a minimal text payload for notifications that do not
// implement SlackNotification.
func (c *SlackChannel) defaultMessage(n contractsnotification.Notification) contractsnotification.SlackMessage {
	return contractsnotification.SlackMessage{
		Text: fmt.Sprintf("New notification: *%T*", n),
	}
}

// ---- internal wire types for the Slack Incoming Webhook API ----

type slackWirePayload struct {
	Text        string                `json:"text,omitempty"`
	Username    string                `json:"username,omitempty"`
	IconEmoji   string                `json:"icon_emoji,omitempty"`
	Channel     string                `json:"channel,omitempty"`
	Attachments []slackWireAttachment `json:"attachments,omitempty"`
}

type slackWireAttachment struct {
	Title  string           `json:"title,omitempty"`
	Text   string           `json:"text,omitempty"`
	Color  string           `json:"color,omitempty"`
	Fields []slackWireField `json:"fields,omitempty"`
	Footer string           `json:"footer,omitempty"`
	Ts     int64            `json:"ts,omitempty"`
}

type slackWireField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// slackPayload converts our public SlackMessage into the Slack wire format.
func slackPayload(msg contractsnotification.SlackMessage) slackWirePayload {
	p := slackWirePayload{
		Text:      msg.Text,
		Username:  msg.Username,
		IconEmoji: msg.IconEmoji,
		Channel:   msg.Channel,
	}
	for _, a := range msg.Attachments {
		wa := slackWireAttachment{
			Title:  a.Title,
			Text:   a.Text,
			Color:  a.Color,
			Footer: a.Footer,
			Ts:     a.Timestamp,
		}
		for _, f := range a.Fields {
			wa.Fields = append(wa.Fields, slackWireField{
				Title: f.Title,
				Value: f.Value,
				Short: f.Short,
			})
		}
		p.Attachments = append(p.Attachments, wa)
	}
	return p
}
