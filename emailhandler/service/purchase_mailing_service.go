package service

import (
	"context"
	"fmt"

	"github.com/domesama/chat-and-notifications/email"
	"github.com/domesama/chat-and-notifications/model"
)

// @@wire-struct@@
type PurchaseMailingService struct {
	EmailInfoService EmailInfoService
	EmailConfig      email.EmailConfig
	EmailSender      email.EmailSender
}

func (p PurchaseMailingService) NotifyPurchaseToShopOwnerEmail(
	ctx context.Context,
	purchase model.PurchaseUpdate) (err error) {

	shopOwnerEmail, err := p.EmailInfoService.GetReceiverEmail(ctx, purchase.ShopOwnerID)
	if err != nil {
		return
	}

	if err = p.EmailSender.SendEmail(ctx, p.generateEmail(purchase, shopOwnerEmail)); err != nil {
		return
	}

	return
}

func (p PurchaseMailingService) generateEmail(
	purchase model.PurchaseUpdate,
	shopOwnerEmailInfo EmailInfo) email.EmailMessage {

	subject := fmt.Sprintf("New purchase order: %s", purchase.ProductName)
	body := fmt.Sprintf(
		"Order ID: %s\nProduct: %s\nAmount: %s %.2f\nStatus: %s\nPurchased: %s",
		purchase.OrderID,
		purchase.ProductName,
		purchase.Currency,
		purchase.Amount,
		purchase.Status,
		purchase.PurchasedAt.Format("2006-01-02 15:04:05"),
	)

	return email.EmailMessage{
		SenderMailAddress:   p.EmailConfig.FromAddress,
		ReceiverMailAddress: shopOwnerEmailInfo.Email,
		Subject:             subject,
		Body:                body,
	}
}
