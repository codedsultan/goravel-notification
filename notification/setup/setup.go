// Package setup contains the Goravel ServiceProvider for codedsultan/goravel-notification.
// Register it in bootstrap/app.go alongside your other providers.
package setup

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"

	"github.com/codedsultan/goravel-notification/channels"
	notificationconsole "github.com/codedsultan/goravel-notification/console"
	"github.com/codedsultan/goravel-notification/notification"
)

// Binding is the service-container key used to resolve the Manager.
// You should not need to reference this directly — use facades.Notification() instead.
const Binding = "goravel.notification"

// ServiceProvider wires codedsultan/goravel-notification into the Goravel application.
//
// Usage in bootstrap/app.go:
//
//	app.Register([]foundation.ServiceProvider{
//	    // ... other providers ...
//	    &notificationsetup.ServiceProvider{},
//	})
type ServiceProvider struct{}

// Register binds the notification Manager into the service container.
// The binding is lazy — the Manager is not constructed until first use.
func (s *ServiceProvider) Register(app foundation.Application) {
	app.BindWith(Binding, func(app foundation.Application, _ map[string]any) (any, error) {
		logger := facades.Log()

		// Queue is optional — notifications that implement ShouldQueue fall back
		// to synchronous delivery when the queue is unavailable.
		var q = facades.Queue()

		mgr := notification.NewManager(logger, q)

		// Register the three built-in channels.
		mgr.Extend(channels.NewMailChannel(facades.Mail(), logger))
		mgr.Extend(channels.NewDatabaseChannel(facades.Orm(), logger))
		mgr.Extend(channels.NewSlackChannel(logger))

		return mgr, nil
	})
}

// Boot registers the artisan commands provided by this package.
func (s *ServiceProvider) Boot(app foundation.Application) {
	facades.Artisan().Register([]console.Command{
		&notificationconsole.NotificationTableCommand{},
	})
}
