// Package console provides artisan commands for codedsultan/goravel-notification.
package console

import (
	"os"
	"path/filepath"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

// NotificationTableCommand generates a migration file that creates the
// notifications table used by the database channel.
//
// Usage:
//
//	./artisan notification:table
//	./artisan migrate
type NotificationTableCommand struct{}

// Signature returns the command signature.
func (c *NotificationTableCommand) Signature() string {
	return "notification:table"
}

// Description returns the command description.
func (c *NotificationTableCommand) Description() string {
	return "Create a migration for the notifications table (database channel)"
}

// Extend returns the command extend configuration.
func (c *NotificationTableCommand) Extend() command.Extend {
	return command.Extend{}
}

// Handle executes the command.
func (c *NotificationTableCommand) Handle(ctx console.Context) error {
	timestamp := time.Now().Format("2006_01_02_150405")
	filename := timestamp + "_create_notifications_table.go"
	dest := filepath.Join("database", "migrations", filename)

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	if _, err := os.Stat(dest); err == nil {
		ctx.Warning("Migration already exists: " + dest)
		return nil
	}

	if err := os.WriteFile(dest, []byte(migrationStub(timestamp)), 0o644); err != nil {
		return err
	}

	ctx.Info("Migration created successfully: " + dest)
	ctx.Info("Run `./artisan migrate` to apply it.")
	return nil
}

func migrationStub(timestamp string) string {
	return `package migrations

import (
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/facades"
)

// CreateNotificationsTable_` + timestamp + ` creates the notifications table
// used by codedsultan/goravel-notification' database channel.
type CreateNotificationsTable_` + timestamp + ` struct{}

func (r *CreateNotificationsTable_` + timestamp + `) Signature() string {
	return "` + timestamp + `_create_notifications_table"
}

func (r *CreateNotificationsTable_` + timestamp + `) Up() error {
	return facades.Schema().Create("notifications", func(table *migration.Blueprint) {
		table.String("id", 36)
		table.Primary("id")
		table.String("type")
		table.String("notifiable_type")
		table.String("notifiable_id")
		table.Text("data")
		table.Timestamp("read_at").Nullable()
		table.Timestamps()
		table.Index("notifiable_type", "notifiable_id")
	})
}

func (r *CreateNotificationsTable_` + timestamp + `) Down() error {
	return facades.Schema().DropIfExists("notifications")
}
`
}
