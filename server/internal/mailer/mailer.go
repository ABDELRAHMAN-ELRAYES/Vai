package mailer

import (
	"context"
	"embed"
)

var (
	SMTPSuccessCode = 250
	MailRetries     = 3
)

// Embed the templates folder inside the GO binary
//
//go:embed templates
var FS embed.FS

type Client interface {
	Send(ctx context.Context, templateFile, toEmail string, data any) (int, error)
}
