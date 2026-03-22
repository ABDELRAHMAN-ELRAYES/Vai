package mailer

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Send: fill the email template with the required data, open connection and then send the email via smtp
func (c *client) Send(ctx context.Context, templateFile, toEmail string, data any) (int, error) {

	// fill the email template with required data
	subject, body, err := c.execute(templateFile, data)
	if err != nil {
		return 0, err
	}

	// send email to reciever (with retries on fial)
	var lastErr error

	for i := 0; i < MailRetries; i++ {

		err := c.deliver(ctx, toEmail, subject, body)
		if err == nil {
			return SMTPSuccessCode, nil
		}

		lastErr = err

		select {
		case <-ctx.Done():
			return 0, ctx.Err()

		case <-time.After(time.Second * time.Duration(i+1)):
		}
	}

	return 0, fmt.Errorf("send mail failed after retries: %w", lastErr)
}

func (c *client) deliver(ctx context.Context, toEmail, subject, htmlBody string) error {
	from := fmt.Sprintf("%s <%s>", c.cfg.FromName, c.cfg.FromAddress)

	var msg strings.Builder
	msg.WriteString("From: " + from + "\r\n") // "\r\n" CRLF - carriage return + Line Feed (required by smtp server not just  \n  to make a new line)
	msg.WriteString("To: " + toEmail + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	cl, err := c.getClient(ctx)
	if err != nil {
		return err
	}

	defer c.releaseClient(cl)

	// Inform the server that the email will be sent from -- and server responds with 250 (ok)
	if err := cl.Mail(c.cfg.FromAddress); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}

	// Inform the server that the email will be sent from -- and server responds with 250 (ok)
	// the server check if valid email, spam rules, ..etc
	if err := cl.Rcpt(toEmail); err != nil {
		return fmt.Errorf("RCPT TO: %w", err)
	}

	// 354 Start mail input and return the mail io writer
	wc, err := cl.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}

	// Write email content and send
	if _, err := wc.Write([]byte(msg.String())); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	// Close the email io writer
	return wc.Close()
}
