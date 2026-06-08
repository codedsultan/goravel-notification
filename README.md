# codedsultan/goravel-notification

[![Go](https://img.shields.io/github/go-mod/go-version/codedsultan/goravel-notification)](https://go.dev/)
[![License](https://img.shields.io/github/license/codedsultan/goravel-notification)](LICENSE)
[![Tests](https://github.com/codedsultan/goravel-notification/actions/workflows/test.yml/badge.svg)](https://github.com/codedsultan/goravel-notification/actions)
[![codecov](https://codecov.io/github/codedsultan/goravel-notification/graph/badge.svg?token=3UQRKBQTKA)](https://codecov.io/github/codedsultan/goravel-notification)

Goravel Notifications package — send alerts via **mail**, **database**, and **Slack**
from a single `Notify()` call. Mirrors Laravel's Notifications system for PHP developers
who want the same ergonomics in Go.

---

## Installation

```bash
go get github.com/codedsultan/goravel-notification
```

Register the service provider in `bootstrap/providers.go`:

```go
import notificationsetup "github.com/codedsultan/goravel-notification/notification/setup"

func Providers() []foundation.ServiceProvider {
    return []foundation.ServiceProvider{
        // ...existing providers...
        &notificationsetup.ServiceProvider{},
    }
}
```

Create the database table (required for the `database` channel):

```bash
./artisan notification:table
./artisan migrate
```

> The `notification:table` command generates a migration file and automatically registers
> it in `bootstrap/migrations.go`.

---

## Quick Start

### 1. Create a notification

```go
// app/notifications/invoice_paid.go
package notifications

import (
    contracts "github.com/codedsultan/goravel-notification/contracts"
    "fmt"
)

type InvoicePaid struct {
    InvoiceID uint
    Amount    float64
}

func NewInvoicePaid(invoiceID uint, amount float64) *InvoicePaid {
    return &InvoicePaid{InvoiceID: invoiceID, Amount: amount}
}

func (n *InvoicePaid) ID() string { return "" } // auto UUID

func (n *InvoicePaid) Via(_ contracts.Notifiable) []string {
    return []string{"mail", "database", "slack"}
}

func (n *InvoicePaid) ToMail(_ contracts.Notifiable) contracts.MailMessage {
    return contracts.MailMessage{
        Subject: "Invoice Paid",
        Content: contracts.MailContent{
            Html: fmt.Sprintf("<p>Your invoice #%d for $%.2f has been paid.</p>", n.InvoiceID, n.Amount),
        },
    }
}

func (n *InvoicePaid) ToDatabase(_ contracts.Notifiable) map[string]any {
    return map[string]any{
        "invoice_id": n.InvoiceID,
        "amount":     fmt.Sprintf("%.2f", n.Amount),
    }
}

func (n *InvoicePaid) ToSlack(_ contracts.Notifiable) contracts.SlackMessage {
    return contracts.SlackMessage{
        Text: fmt.Sprintf("Invoice #%d for $%.2f has been paid.", n.InvoiceID, n.Amount),
        Attachments: []contracts.SlackAttachment{
            {
                Title: "Invoice Paid",
                Text:  fmt.Sprintf("Amount: $%.2f", n.Amount),
                Color: "good",
            },
        },
    }
}
```

### 2. Make your model Notifiable

```go
// app/models/user.go
package models

import (
    "fmt"
    "github.com/goravel/framework/database/orm"
    notificationfacades "github.com/codedsultan/goravel-notification/facades"
    contracts "github.com/codedsultan/goravel-notification/contracts"
)

type User struct {
    orm.Model
    Name            string
    Email           string
    SlackWebhookURL string
}

// RouteNotificationFor satisfies contracts.Notifiable
func (u *User) RouteNotificationFor(channel string) string {
    switch channel {
    case "mail":     return u.Email
    case "slack":    return u.SlackWebhookURL
    case "database": return fmt.Sprintf("%d", u.ID)
    }
    return ""
}

// Notify is a convenience helper (mirrors Laravel's Notifiable trait)
func (u *User) Notify(n contracts.Notification) error {
    return notificationfacades.Notification().Send(u, n)
}
```

### 3. Send it

```go
// Option A — via the model helper
_ = user.Notify(notifications.NewInvoicePaid(42, 99.99))

// Option B — via the facade directly
_ = notificationfacades.Notification().Send(&user, notifications.NewInvoicePaid(42, 99.99))

// Option C — always synchronous, even for ShouldQueue notifications
_ = notificationfacades.Notification().SendNow(&user, notifications.NewInvoicePaid(42, 99.99))
```

---

## Channels

| Name       | Driver                | What it needs                                           |
|------------|-----------------------|---------------------------------------------------------|
| `mail`     | `facades.Mail()`      | `RouteNotificationFor("mail")` → email address         |
| `database` | `facades.Orm()`       | `RouteNotificationFor("database")` → model's string PK |
| `slack`    | HTTP incoming webhook | `RouteNotificationFor("slack")` → webhook URL          |

### Mail

Uses Goravel's mail facade. Set `Content.Html` for HTML emails or `Content.View` for
a Goravel view template. Do **not** use `Content.Text` — the framework treats it as a
template path, not a raw string.

### Database

Stores notifications in the `notifications` table. The `data` column is JSON-encoded
from the map returned by `ToDatabase()`.

### Slack

Posts to a Slack incoming webhook URL. Set `RouteNotificationFor("slack")` to return
the webhook URL on your notifiable model.

### Custom channels

```go
type SMSChannel struct{}

func (SMSChannel) Name() string { return "sms" }
func (SMSChannel) Send(notifiable contracts.Notifiable, n contracts.Notification) error {
    phone := notifiable.RouteNotificationFor("sms")
    // call your SMS provider with phone
    return nil
}

// Register in a ServiceProvider Boot():
notificationfacades.Notification().Extend(&SMSChannel{})
```

---

## Queued Notifications

Implement `contracts.ShouldQueue` to dispatch via Goravel's queue:

```go
func (n *InvoicePaid) OnQueue() string      { return "notifications" }
func (n *InvoicePaid) OnConnection() string { return "" }
```

Use `SendNow()` to bypass the queue when needed.

---

## Testing

```go
import mocksnotification "github.com/codedsultan/goravel-notification/mocks/notification"

func TestSomething(t *testing.T) {
    mock := mocksnotification.NewManager(t)
    mock.On("Send", mock.Anything, mock.Anything).Return(nil)

    // inject mock into your service
}
```

---

## License

MIT — see [LICENSE](LICENSE).

