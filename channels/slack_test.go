package channels_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/codedsultan/goravel-notification/channels"
	contractsnotification "github.com/codedsultan/goravel-notification/contracts"
	mocklog "github.com/codedsultan/goravel-notification/mocks/log"
)

// ---- Fakes ----

type slackNotifiable struct{ webhook string }

func (s *slackNotifiable) RouteNotificationFor(channel string) string {
	if channel == "slack" {
		return s.webhook
	}
	return ""
}

// plainSlackNotification falls back to the default text payload.
type plainSlackNotification struct{}

func (p *plainSlackNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{"slack"}
}
func (p *plainSlackNotification) ID() string { return "" }

// richSlackNotification implements SlackNotification.
type richSlackNotification struct{}

func (r *richSlackNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{"slack"}
}
func (r *richSlackNotification) ID() string { return "" }
func (r *richSlackNotification) ToSlack(_ contractsnotification.Notifiable) contractsnotification.SlackMessage {
	return contractsnotification.SlackMessage{
		Text:      "Deployment succeeded :rocket:",
		IconEmoji: ":white_check_mark:",
		Attachments: []contractsnotification.SlackAttachment{
			{
				Title: "Details",
				Text:  "Branch: main • SHA: abc1234",
				Color: "good",
				Fields: []contractsnotification.SlackField{
					{Title: "Environment", Value: "production", Short: true},
					{Title: "Duration", Value: "42s", Short: true},
				},
			},
		},
	}
}

// ---- Tests ----

func TestSlackChannel_Name(t *testing.T) {
	ch := channels.NewSlackChannel(nil)
	assert.Equal(t, "slack", ch.Name())
}

func TestSlackChannel_Send_PostsToWebhook_DefaultMessage(t *testing.T) {
	logger := mocklog.NewMockLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ch := channels.NewSlackChannel(logger)
	notifiable := &slackNotifiable{webhook: server.URL}
	n := &plainSlackNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	assert.Contains(t, received["text"], "plainSlackNotification")
}

func TestSlackChannel_Send_PostsToWebhook_RichMessage(t *testing.T) {
	logger := mocklog.NewMockLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ch := channels.NewSlackChannel(logger)
	notifiable := &slackNotifiable{webhook: server.URL}
	n := &richSlackNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	assert.Equal(t, "Deployment succeeded :rocket:", received["text"])
	assert.Equal(t, ":white_check_mark:", received["icon_emoji"])
}

func TestSlackChannel_Send_ReturnsError_WhenEmptyWebhook(t *testing.T) {
	logger := mocklog.NewMockLog(t)
	ch := channels.NewSlackChannel(logger)

	err := ch.Send(&slackNotifiable{webhook: ""}, &plainSlackNotification{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty webhook URL")
}

func TestSlackChannel_Send_ReturnsError_WhenNon2xxResponse(t *testing.T) {
	logger := mocklog.NewMockLog(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	ch := channels.NewSlackChannel(logger)
	err := ch.Send(&slackNotifiable{webhook: server.URL}, &plainSlackNotification{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-2xx")
}
