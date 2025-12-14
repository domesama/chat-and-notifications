package generalnotifications

import (
	"time"
)

// NotificationEnvelope is the wrapper structure sent over WebSocket
// Frontend can parse the payload field based on notification_type
type NotificationEnvelope struct {
	NotificationType NotificationType `json:"notification_type"`
	Timestamp        time.Time        `json:"timestamp"`
	Payload          any              `json:"payload"`
}

// NewNotificationEnvelope creates a new envelope with auto-generated ID and timestamp
func NewNotificationEnvelope(notificationType NotificationType,
	payload interface{}) NotificationEnvelope {
	return NotificationEnvelope{
		NotificationType: notificationType,
		Timestamp:        time.Now(),
		Payload:          payload,
	}
}
