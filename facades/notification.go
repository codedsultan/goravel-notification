// Package facades exposes the Notification singleton via the standard Goravel
// facade pattern. Import this package and call facades.Notification() anywhere
// in your application.
package facades

import (
	"log"

	contractsnotification "github.com/codedsultan/goravel-notification/contracts"
	"github.com/codedsultan/goravel-notification/notification/setup"

	"github.com/goravel/framework/facades"
)

// Notification returns the Manager singleton from the service container.
// It panics on startup if the ServiceProvider was not registered — this mirrors
// the behaviour of other Goravel facades (facades.Mail, facades.Queue, etc.)
// and surfaces misconfiguration immediately rather than at call time.
func Notification() contractsnotification.Manager {
	instance, err := facades.App().MakeWith(setup.Binding, map[string]any{})
	if err != nil {
		log.Panicf(
			"notifications: failed to resolve Manager from container (did you register setup.ServiceProvider?): %v",
			err,
		)
	}

	mgr, ok := instance.(contractsnotification.Manager)
	if !ok {
		log.Panicf("notifications: container binding %q is not a contracts.Manager", setup.Binding)
	}

	return mgr
}
