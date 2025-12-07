package email

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

type EmailService interface {
	SendEmail(params *SendEmailParams) (*SendEmailResponse, error)
}
