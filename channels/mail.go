// Package channels contains the built-in delivery drivers for codedsultan/goravel-notification.
package channels

import (
	"fmt"

	contractsnotification "github.com/codedsultan/goravel-notification/contracts"

	"github.com/goravel/framework/contracts/log"
	contractsmail "github.com/goravel/framework/contracts/mail"
)

// MailChannel delivers notifications via Goravel's mail facade.
// It respects the full contracts/mail.Mail interface — Html, plain-text, views,
// attachments, and custom headers are all forwarded from MailMessage.
type MailChannel struct {
	mail contractsmail.Mail
	log  log.Log
}

// NewMailChannel creates a MailChannel. The mail argument is facades.Mail()
// injected by the ServiceProvider, making the channel fully testable.
func NewMailChannel(mail contractsmail.Mail, logger log.Log) *MailChannel {
	return &MailChannel{mail: mail, log: logger}
}

// Name satisfies contracts.Channel.
func (c *MailChannel) Name() string { return "mail" }

// Send satisfies contracts.Channel.
// It builds a Mailable from the notification's ToMail() output (or a sensible
// default if the notification does not implement MailableNotification), then
// hands it to the framework's mail driver.
func (c *MailChannel) Send(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	// Resolve the recipient address from the notifiable.
	to := notifiable.RouteNotificationFor("mail")
	if to == "" {
		return fmt.Errorf("mail channel: %T.RouteNotificationFor(\"mail\") returned empty address", notifiable)
	}

	// Build the MailMessage — use ToMail() if available, else fall back.
	var msg contractsnotification.MailMessage
	if mn, ok := n.(contractsnotification.MailableNotification); ok {
		msg = mn.ToMail(notifiable)
	} else {
		msg = c.defaultMessage(n)
	}

	// Override recipient if the message specifies one explicitly.
	recipients := []string{to}
	if msg.To != "" {
		recipients = []string{msg.To}
	}

	subject := msg.Subject
	if subject == "" {
		subject = fmt.Sprintf("Notification: %T", n)
	}

	// Build the mailable and send.
	mailable := &NotificationMailable{
		envelope: &contractsmail.Envelope{
			To:      recipients,
			Subject: subject,
		},
		content: &contractsmail.Content{
			Html: msg.Content.Html,
			Text: msg.Content.Text,
			View: msg.Content.View,
			With: msg.Content.With,
		},
		attachments: msg.Attachments,
		headers:     msg.Headers,
	}

	if msg.From != "" {
		mailable.envelope.From = contractsmail.Address{Address: msg.From}
	}

	if err := c.mail.Send(mailable); err != nil {
		return fmt.Errorf("mail channel: failed to send %T: %w", n, err)
	}
	return nil
}

// defaultMessage builds a plain-text fallback when the notification does not
// implement MailableNotification.
func (c *MailChannel) defaultMessage(n contractsnotification.Notification) contractsnotification.MailMessage {
	return contractsnotification.MailMessage{
		Subject: fmt.Sprintf("Notification: %T", n),
		Content: contractsnotification.MailContent{
			Text: fmt.Sprintf("You have a new %T notification.", n),
		},
	}
}

// NotificationMailable adapts our MailMessage into the contractsmail.Mailable
// interface expected by the goravel/framework mail driver.
type NotificationMailable struct {
	envelope    *contractsmail.Envelope
	content     *contractsmail.Content
	attachments []string
	headers     map[string]string
}

// Envelope returns the mail envelope.
func (m *NotificationMailable) Envelope() *contractsmail.Envelope {
	return m.envelope
}

// Content returns the mail content.
func (m *NotificationMailable) Content() *contractsmail.Content {
	return m.content
}

// Attachments returns the mail attachments.
func (m *NotificationMailable) Attachments() []string {
	return m.attachments
}

// Headers returns custom headers.
func (m *NotificationMailable) Headers() map[string]string {
	return m.headers
}

// Queue returns queue configuration (nil for immediate sending).
func (m *NotificationMailable) Queue() *contractsmail.Queue {
	return nil
}
