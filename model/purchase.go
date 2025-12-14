package model

import "time"

// PurchaseUpdate represents the payload for purchase notifications & email
type PurchaseUpdate struct {
	OrderID     string    `json:"order_id" bson:"_id" binding:"required"`
	BuyerID     string    `json:"buyer_id" bson:"buyer_id" binding:"required"`
	ShopOwnerID string    `json:"shop_owner_id" bson:"shop_owner_id" binding:"required"`
	ProductName string    `json:"product_name" bson:"product_name" binding:"required"`
	Amount      float64   `json:"amount" bson:"amount" binding:"required"`
	Currency    string    `json:"currency" bson:"currency" binding:"required"`
	Status      string    `json:"status" bson:"status" binding:"required"`
	PurchasedAt time.Time `json:"purchased_at" bson:"purchased_at"`
}
