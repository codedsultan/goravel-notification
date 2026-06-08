package notification

import (
	contractsnotification "github.com/codedsultan/goravel-notification/contracts"
)

// sendNotificationJob is an internal queue.Job that re-delivers a notification
// synchronously when consumed by a Goravel queue worker.
// It is created automatically by Manager.dispatchQueued and never used directly.
type sendNotificationJob struct {
	manager    *Manager
	notifiable contractsnotification.Notifiable
	n          contractsnotification.Notification
}

// Signature uniquely identifies this job type in Goravel's job registry.
func (j *sendNotificationJob) Signature() string {
	return "goravel_notifications:send"
}

// Handle is called by the queue worker. It delegates back to the synchronous
// dispatch path, which iterates over channels and calls Send on each driver.
func (j *sendNotificationJob) Handle(_ ...any) error {
	return j.manager.dispatchSync(j.notifiable, j.n)
}
