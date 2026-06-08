package channels_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/codedsultan/goravel-notification/channels"
	contractsnotification "github.com/codedsultan/goravel-notification/contracts"
	mocklog "github.com/codedsultan/goravel-notification/mocks/log"
	mockmail "github.com/codedsultan/goravel-notification/mocks/mail"
	contractsmail "github.com/goravel/framework/contracts/mail"
)

// ---- Fakes ----

type mailNotifiable struct{ addr string }

func (m *mailNotifiable) RouteNotificationFor(channel string) string {
	if channel == "mail" {
		return m.addr
	}
	return ""
}

// plainNotification does NOT implement MailableNotification — tests the fallback path.
type plainNotification struct{}

func (p *plainNotification) Via(_ contractsnotification.Notifiable) []string { return []string{"mail"} }
func (p *plainNotification) ID() string                                      { return "" }

// richNotification implements MailableNotification — tests the ToMail() path.
type richNotification struct {
	msg contractsnotification.MailMessage
}

func (r *richNotification) Via(_ contractsnotification.Notifiable) []string { return []string{"mail"} }
func (r *richNotification) ID() string                                      { return "" }
func (r *richNotification) ToMail(_ contractsnotification.Notifiable) contractsnotification.MailMessage {
	return r.msg
}

// ---- Tests ----

func TestMailChannel_Name(t *testing.T) {
	ch := channels.NewMailChannel(nil, nil)
	assert.Equal(t, "mail", ch.Name())
}

func TestMailChannel_Send_UsesDefaultMessage_WhenNotMailableNotification(t *testing.T) {
	logger := mocklog.NewMockLog(t)
	mailer := mockmail.NewMockMail(t)

	// The channel should call mailer.Send with a Mailable — capture it.
	mailer.On("Send", mock.AnythingOfType("*channels.NotificationMailable")).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	mailer.AssertExpectations(t)
}

func TestMailChannel_Send_UsesToMail_WhenMailableNotification(t *testing.T) {
	logger := mocklog.NewMockLog(t)
	mailer := mockmail.NewMockMail(t)

	mailer.On("Send", mock.AnythingOfType("*channels.NotificationMailable")).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &richNotification{
		msg: contractsnotification.MailMessage{
			Subject: "Invoice Paid",
			Content: contractsnotification.MailContent{Text: "Your invoice was paid."},
		},
	}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	mailer.AssertExpectations(t)
}

func TestMailChannel_Send_ReturnsError_WhenEmptyAddress(t *testing.T) {
	logger := mocklog.NewMockLog(t)
	mailer := mockmail.NewMockMail(t)

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailNotifiable{addr: ""} // no address
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty address")
	mailer.AssertNotCalled(t, "Send", mock.Anything)
}

func TestMailChannel_Send_WrapsMailerError(t *testing.T) {
	logger := mocklog.NewMockLog(t)
	mailer := mockmail.NewMockMail(t)

	mailer.On("Send", mock.Anything).
		Return(errors.New("SMTP connection refused")).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP connection refused")
}

// Verify the Mailable adapter satisfies the contractsmail.Mailable interface at compile time.
var _ contractsmail.Mailable = (*channels.NotificationMailable)(nil)
