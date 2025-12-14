package stub

import (
	"time"

	"github.com/domesama/chat-and-notifications/model"
)

func CreatePurchaseNotificationPayload(orderID, productName string,
	amount float64) model.PurchaseUpdate {
	return model.PurchaseUpdate{
		OrderID:     orderID,
		BuyerID:     "buyer-default",
		ShopOwnerID: "shop-owner-default",
		ProductName: productName,
		Amount:      amount,
		Currency:    "USD",
		Status:      "confirmed",
		PurchasedAt: time.Now(),
	}
}

func CreatePurchaseUpdate(orderID, buyerID, shopOwnerID, productName string,
	amount float64) model.PurchaseUpdate {
	return model.PurchaseUpdate{
		OrderID:     orderID,
		BuyerID:     buyerID,
		ShopOwnerID: shopOwnerID,
		ProductName: productName,
		Amount:      amount,
		Currency:    "USD",
		Status:      "confirmed",
		PurchasedAt: time.Now(),
	}
}

func CreatePaymentReminderNotificationPayload(orderID string, amount float64, daysOverdue int,
	message string) model.PaymentReminder {
	return model.PaymentReminder{
		OrderID:     orderID,
		Amount:      amount,
		Currency:    "USD",
		DueDate:     time.Now().Add(7 * 24 * time.Hour),
		DaysOverdue: daysOverdue,
		Message:     message,
	}
}

func CreateShippingUpdateNotificationPayload(orderID, trackingNumber, status, location string) model.ShippingUpdate {
	return model.ShippingUpdate{
		OrderID:           orderID,
		TrackingNumber:    trackingNumber,
		Status:            status,
		Location:          location,
		EstimatedDelivery: time.Now().Add(48 * time.Hour),
		UpdatedAt:         time.Now(),
	}
}
