package service

import (
	"context"

	"github.com/domesama/chat-and-notifications/email"
	"github.com/domesama/chat-and-notifications/model"
)

// @@wire-struct@@
type ChatMailingService struct {
	EmailInfoService EmailInfoService
	EmailConfig      email.EmailConfig
	EmailSender      email.EmailSender
}

func (c ChatMailingService) NotifyChatMessageToReceiverEmail(
	ctx context.Context,
	message model.ChatMessage) (err error) {

	receiverEmail, err := c.EmailInfoService.GetReceiverEmail(ctx, message.ReceiverID)
	if err != nil {
		return
	}

	if err = c.EmailSender.SendEmail(ctx, c.generateEmail(message, receiverEmail)); err != nil {
		return
	}

	return
}

func (c ChatMailingService) generateEmail(
	chatMessage model.ChatMessage,
	receiverEmailInfo EmailInfo) email.EmailMessage {

	return email.EmailMessage{
		SenderMailAddress:   c.EmailConfig.FromAddress,
		ReceiverMailAddress: receiverEmailInfo.Email,
		Subject:             "You have a new chat message from " + receiverEmailInfo.Name,
		Body:                chatMessage.Content,
	}
}
