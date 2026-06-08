# codedsultan/goravel-notification

[![Go](https://img.shields.io/github/go-mod/go-version/codedsultan/goravel-notification)](https://go.dev/)
[![License](https://img.shields.io/github/license/codedsultan/goravel-notification)](LICENSE)
[![Tests](https://github.com/codedsultan/goravel-notification/actions/workflows/test.yml/badge.svg)](https://github.com/codedsultan/goravel-notification/actions)

Official Goravel Notifications package — send alerts via **mail**, **database**, and **Slack**
from a single `Notify()` call. Mirrors Laravel's Notifications system for PHP developers
who want the same ergonomics in Go.

---

## Installation

```bash
go get github.com/codedsultan/goravel-notification
```

Register the service provider in `bootstrap/app.go`:

```go
import notificationsetup "github.com/codedsultan/goravel-notification/notification/setup"

app.Register([]foundation.ServiceProvider{
    // ...existing providers...
    &notificationsetup.ServiceProvider{},
})
```

Create the database table (required for the `database` channel):

```bash
./artisan notification:table
./artisan migrate
```

---

## Quick Start

### 1. Create a notification

```go
// app/notifications/invoice_paid.go
package notifications

import (
    contractsnotification "github.com/codedsultan/goravel-notification/contracts"
    "myapp/app/models"
    "fmt"
)

type InvoicePaid struct {
    Invoice *models.Invoice
}

func NewInvoicePaid(invoice *models.Invoice) *InvoicePaid {
    return &InvoicePaid{Invoice: invoice}
}

func (n *InvoicePaid) ID() string { return "" } // auto UUID

func (n *InvoicePaid) Via(_ contractsnotification.Notifiable) []string {
    return []string{"mail", "database"}
}

func (n *InvoicePaid) ToMail(_ contractsnotification.Notifiable) contractsnotification.MailMessage {
    return contractsnotification.MailMessage{
        Subject: "Invoice Paid",
        Content: contractsnotification.MailContent{
            Text: fmt.Sprintf("Your invoice #%d for $%.2f has been paid.", n.Invoice.ID, n.Invoice.Amount),
        },
    }
}

func (n *InvoicePaid) ToDatabase(_ contractsnotification.Notifiable) map[string]any {
    return map[string]any{
        "invoice_id": n.Invoice.ID,
        "amount":     n.Invoice.Amount,
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
    contractsnotification "github.com/codedsultan/goravel-notification/contracts"
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
func (u *User) Notify(n contractsnotification.Notification) error {
    return notificationfacades.Notification().Send(u, n)
}
```

### 3. Send it

```go
// In a controller, service, or job:

// Option A — via the model helper
_ = user.Notify(notifications.NewInvoicePaid(invoice))

// Option B — via the facade directly
_ = facades.Notification().Send(&user, notifications.NewInvoicePaid(invoice))

// Option C — always synchronous, even for ShouldQueue notifications
_ = facades.Notification().SendNow(&user, notifications.NewInvoicePaid(invoice))
```

---

## Channels

| Name       | Driver                 | What it needs                                              |
|------------|------------------------|------------------------------------------------------------|
| `mail`     | `facades.Mail()`       | `RouteNotificationFor("mail")` → email address            |
| `database` | `facades.Orm()`        | `RouteNotificationFor("database")` → model's string PK    |
| `slack`    | HTTP incoming webhook  | `RouteNotificationFor("slack")` → webhook URL             |

### Custom channels

```go
type SMSChannel struct{}

func (SMSChannel) Name() string { return "sms" }
func (SMSChannel) Send(notifiable contracts.Notifiable, n contracts.Notification) error {
    // call your SMS provider here
    return nil
}

// In a ServiceProvider Boot():
facades.Notification().Extend(&SMSChannel{})
```

---

## Queued Notifications

Implement `contracts.ShouldQueue` to dispatch via Goravel's queue:

```go
func (n *InvoicePaid) OnQueue() string     { return "notifications" }
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
