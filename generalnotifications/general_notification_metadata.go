package generalnotifications

import (
	"github.com/domesama/chat-and-notifications/websocket"
)

// NotificationMetadata contains the routing information for general notifications
type NotificationMetadata struct {
	UserID string `json:"user_id" bson:"user_id" form:"user_id" binding:"required"`
}

func (n NotificationMetadata) ToWebSocketMetadata() websocket.Metadata {
	return websocket.Metadata{
		"user_id": {n.UserID},
	}
}
