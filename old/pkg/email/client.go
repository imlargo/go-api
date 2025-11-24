package email

import (
	"github.com/resend/resend-go/v2"
)

type EmailClient interface {
	SendEmail(params *SendEmailParams) (*SendEmailResponse, error)
}

type SendEmailParams struct {
	From    string
	To      []string
	Subject string
	Html    string
	Text    string
	Cc      []string
	Bcc     []string
	ReplyTo string
}

type SendEmailResponse struct {
	ID string
}

type emailClient struct {
	client *resend.Client
}

func NewEmailClient(apiKey string) EmailClient {
	client := resend.NewClient(apiKey)
	return &emailClient{
		client: client,
	}
}

func (e *emailClient) SendEmail(params *SendEmailParams) (*SendEmailResponse, error) {
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

	return &SendEmailResponse{
		ID: sent.Id,
	}, nil
}
