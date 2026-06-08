package channels_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/codedsultan/goravel-notification/channels"
	contractsnotification "github.com/codedsultan/goravel-notification/contracts"
	mocklog "github.com/codedsultan/goravel-notification/mocks/log"
	mockorm "github.com/codedsultan/goravel-notification/mocks/orm"
	mockquery "github.com/codedsultan/goravel-notification/mocks/orm/query"
)

// ---- Fakes ----

type dbNotifiable struct{ id string }

func (d *dbNotifiable) RouteNotificationFor(channel string) string {
	if channel == "database" {
		return d.id
	}
	return ""
}

// dbNotification does NOT implement DatabaseNotification — tests the fallback payload.
type dbNotification struct{}

func (d *dbNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{"database"}
}
func (d *dbNotification) ID() string { return "fixed-uuid-1234" }

// richDbNotification implements DatabaseNotification.
type richDbNotification struct{}

func (r *richDbNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{"database"}
}
func (r *richDbNotification) ID() string { return "" }
func (r *richDbNotification) ToDatabase(_ contractsnotification.Notifiable) map[string]any {
	return map[string]any{"invoice_id": 99, "amount": "250.00"}
}

// ---- Tests ----

func TestDatabaseChannel_Name(t *testing.T) {
	ch := channels.NewDatabaseChannel(nil, nil)
	assert.Equal(t, "database", ch.Name())
}

func TestDatabaseChannel_Send_InsertsRecord_WithDefaultPayload(t *testing.T) {
	logger := mocklog.NewMockLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	query := mockquery.NewMockQuery(t)
	query.On("Create", mock.AnythingOfType("*channels.DatabaseNotificationModel")).
		Return(nil).Once()

	o := mockorm.NewMockOrm(t)
	o.On("Query").Return(query)

	ch := channels.NewDatabaseChannel(o, logger)
	notifiable := &dbNotifiable{id: "42"}
	n := &dbNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	query.AssertExpectations(t)
}

func TestDatabaseChannel_Send_InsertsRecord_WithCustomPayload(t *testing.T) {
	logger := mocklog.NewMockLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	query := mockquery.NewMockQuery(t)
	query.On("Create", mock.MatchedBy(func(r *channels.DatabaseNotificationModel) bool {
		// Verify the JSON payload contains the custom fields.
		return r.NotifiableID == "42" && len(r.Data) > 0
	})).Return(nil).Once()

	o := mockorm.NewMockOrm(t)
	o.On("Query").Return(query)

	ch := channels.NewDatabaseChannel(o, logger)
	notifiable := &dbNotifiable{id: "42"}
	n := &richDbNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	query.AssertExpectations(t)
}

func TestDatabaseChannel_Send_ReturnsError_WhenEmptyID(t *testing.T) {
	logger := mocklog.NewMockLog(t)
	o := mockorm.NewMockOrm(t)

	ch := channels.NewDatabaseChannel(o, logger)
	notifiable := &dbNotifiable{id: ""} // no routing ID
	n := &dbNotification{}

	err := ch.Send(notifiable, n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty ID")
}

func TestDatabaseChannel_Send_WrapsOrmError(t *testing.T) {
	logger := mocklog.NewMockLog(t)

	query := mockquery.NewMockQuery(t)
	query.On("Create", mock.Anything).Return(errors.New("unique constraint violation")).Once()

	o := mockorm.NewMockOrm(t)
	o.On("Query").Return(query)

	ch := channels.NewDatabaseChannel(o, logger)
	err := ch.Send(&dbNotifiable{id: "1"}, &dbNotification{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unique constraint violation")
}
