package email

type EmailMessage struct {
	SenderMailAddress   string `json:"sender_mail_address"`
	ReceiverMailAddress string `json:"receiver_mail_address"`
	Subject             string `json:"subject"`
	Body                string `json:"body"`
}
