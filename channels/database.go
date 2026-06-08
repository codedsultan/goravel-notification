package channels

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	contractsnotification "github.com/codedsultan/goravel-notification/contracts"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/log"
)

// DatabaseNotificationModel is the ORM model written to the notifications table.
// It mirrors Laravel's database notification schema exactly, making migration
// from PHP/Laravel codebases straightforward.
//
// Run `./artisan notification:table && ./artisan migrate` to create the table.
type DatabaseNotificationModel struct {
	// ID is a UUID string primary key — avoids auto-increment races in distributed systems.
	ID string `gorm:"primaryKey;type:varchar(36);column:id"`

	// Type is the fully-qualified Go type name of the notification struct.
	// e.g. "github.com/myapp/app/notifications.InvoicePaid"
	Type string `gorm:"not null;column:type"`

	// NotifiableType is the fully-qualified Go type name of the notifiable model.
	NotifiableType string `gorm:"not null;column:notifiable_type"`

	// NotifiableID is the notifiable's routing key (its string primary key).
	NotifiableID string `gorm:"not null;column:notifiable_id"`

	// Data is the JSON-encoded payload returned by ToDatabase() or the default map.
	Data string `gorm:"type:text;not null;column:data"`

	// ReadAt is null until the notification is marked as read.
	ReadAt *time.Time `gorm:"column:read_at"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName tells GORM to use "notifications" instead of the default plural of the struct name.
func (DatabaseNotificationModel) TableName() string { return "notifications" }

// DatabaseChannel persists notifications to the database so they can be
// retrieved, displayed, and marked as read in the UI.
type DatabaseChannel struct {
	orm orm.Orm
	log log.Log
}

// NewDatabaseChannel creates a DatabaseChannel.
// The orm argument is facades.Orm() injected by the ServiceProvider.
func NewDatabaseChannel(o orm.Orm, logger log.Log) *DatabaseChannel {
	return &DatabaseChannel{orm: o, log: logger}
}

// Name satisfies contracts.Channel.
func (c *DatabaseChannel) Name() string { return "database" }

// Send satisfies contracts.Channel.
// It marshals the notification payload and inserts a row into the notifications table.
func (c *DatabaseChannel) Send(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	notifiableID := notifiable.RouteNotificationFor("database")
	if notifiableID == "" {
		return fmt.Errorf("database channel: %T.RouteNotificationFor(\"database\") returned empty ID", notifiable)
	}

	// Resolve the data payload.
	var payload map[string]any
	if dn, ok := n.(contractsnotification.DatabaseNotification); ok {
		payload = dn.ToDatabase(notifiable)
	} else {
		payload = map[string]any{
			"type": fmt.Sprintf("%T", n),
		}
	}

	dataJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("database channel: failed to marshal payload for %T: %w", n, err)
	}

	// Resolve the record ID.
	id := n.ID()
	if id == "" {
		id = uuid.NewString()
	}

	record := &DatabaseNotificationModel{
		ID:             id,
		Type:           fmt.Sprintf("%T", n),
		NotifiableType: fmt.Sprintf("%T", notifiable),
		NotifiableID:   notifiableID,
		Data:           string(dataJSON),
	}

	if err := c.orm.Query().Create(record); err != nil {
		return fmt.Errorf("database channel: failed to insert notification record: %w", err)
	}

	c.log.Debugf("notifications: persisted %T to database (id=%s)", n, id)
	return nil
}
