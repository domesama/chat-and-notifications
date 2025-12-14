package generalnotifications

type NotificationType string

const (
	ChatNotification            NotificationType = "chat_notification"
	PurchaseNotification        NotificationType = "purchase_notification"
	PaymentReminderNotification NotificationType = "payment_reminder_notification"
	ShippingUpdateNotification  NotificationType = "shipping_update_notification"
)
