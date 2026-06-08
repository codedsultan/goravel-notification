// Package config provides configuration structures for notifications.
package config

import "github.com/goravel/framework/facades"

// init registers the notification configuration block with Goravel's config system.
// Values are read from environment variables, with the defaults shown below.
//
// Add this file to your application's config/ directory and import it in
// bootstrap/app.go to enable configuration overrides.
func init() {
	config := facades.Config()

	config.Add("notification", map[string]any{
		// default is the fallback channel used when Via() is not implemented.
		// In practice Via() is always required, so this is advisory only.
		"default": config.Env("NOTIFICATION_CHANNEL", "mail"),

		"channels": map[string]any{
			// slack contains settings for the Slack incoming-webhook channel.
			"slack": map[string]any{
				// webhook is the default Slack webhook URL.
				// Per-notifiable routing is done via RouteNotificationFor("slack").
				"webhook": config.Env("SLACK_WEBHOOK_URL", ""),
				// username overrides the bot name in Slack messages.
				"username": config.Env("SLACK_USERNAME", "Goravel"),
				// icon_emoji overrides the bot icon.
				"icon_emoji": config.Env("SLACK_ICON_EMOJI", ":bell:"),
			},

			// database contains settings for the database channel.
			"database": map[string]any{
				// connection is the DB connection name (empty = default connection).
				"connection": config.Env("NOTIFICATION_DB_CONNECTION", ""),
			},
		},
	})
}
