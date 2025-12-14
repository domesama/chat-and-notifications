package model

import "time"

// PaymentReminder represents the payload for payment reminder
type PaymentReminder struct {
	OrderID     string    `json:"order_id" binding:"required"`
	Amount      float64   `json:"amount" binding:"required"`
	Currency    string    `json:"currency" binding:"required"`
	DueDate     time.Time `json:"due_date" binding:"required"`
	DaysOverdue int       `json:"days_overdue"`
	Message     string    `json:"message"`
}
