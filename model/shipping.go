package model

import "time"

// ShippingUpdate represents the payload for shipping update notifications
type ShippingUpdate struct {
	OrderID           string    `json:"order_id" binding:"required"`
	TrackingNumber    string    `json:"tracking_number"`
	Status            string    `json:"status" binding:"required"`
	Location          string    `json:"location"`
	EstimatedDelivery time.Time `json:"estimated_delivery"`
	UpdatedAt         time.Time `json:"updated_at"`
}
