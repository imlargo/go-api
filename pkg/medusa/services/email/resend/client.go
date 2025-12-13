package email

import (
	"github.com/imlargo/go-api/pkg/medusa/services/email"
	"github.com/resend/resend-go/v2"
)

type resendEmailClient struct {
	client *resend.Client
}

func NewResendEmailClient(apiKey string) email.EmailService {
	client := resend.NewClient(apiKey)
	return &resendEmailClient{
		client: client,
	}
}

func (e *resendEmailClient) SendEmail(params *email.SendEmailParams) (*email.SendEmailResponse, error) {

	sendParams := &resend.SendEmailRequest{
		From:    params.From,
		To:      params.To,
		Subject: params.Subject,
		Html:    params.Html,
		Text:    params.Text,
	}

	if len(params.Cc) > 0 {
		sendParams.Cc = params.Cc
	}

	if len(params.Bcc) > 0 {
		sendParams.Bcc = params.Bcc
	}

	if params.ReplyTo != "" {
		sendParams.ReplyTo = params.ReplyTo
	}

	sent, err := e.client.Emails.Send(sendParams)
	if err != nil {
		return nil, err
	}

	return &email.SendEmailResponse{
		ID: sent.Id,
	}, nil
}
