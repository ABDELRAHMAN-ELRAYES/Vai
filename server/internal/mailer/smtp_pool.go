package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
)

// Reuse the connection on the pool channel
func (c *client) getClient(ctx context.Context) (*smtp.Client, error) {

	select {
	// if the pool channel is not empty(have available connections -> return one)
	case cl := <-c.pool:
		return cl, nil
	default:
		return c.newSMTPClient(ctx)

	}
}

func (c *client) newSMTPClient(ctx context.Context) (*smtp.Client, error) {

	addr := fmt.Sprintf("%s:%d", c.cfg.SMTPHost, c.cfg.SMTPPort)

	// Open a TCP connection to the SMTP server
	d := net.Dialer{}

	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	// Create SMTP Client (Vai)(wrapping the TCP connection to the smtp server)
	cl, err := smtp.NewClient(conn, c.cfg.SMTPHost)
	if err != nil {
		return nil, fmt.Errorf("smtp client: %w", err)
	}

	// Upgrade the connection to encrypted TLS
	err = cl.StartTLS(&tls.Config{
		ServerName: c.cfg.SMTPHost,
		MinVersion: tls.VersionTLS12,
	})

	if err != nil {
		return nil, fmt.Errorf("starttls: %w", err)
	}

	auth := smtp.PlainAuth("", c.cfg.SMTPUser, c.cfg.SMTPPassword, c.cfg.SMTPHost)

	// Authenticate Vai to the SMTP server
	if err := cl.Auth(auth); err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	return cl, nil
}

func (c *client) releaseClient(cl *smtp.Client) {

	select {
	// Send the client (connection) to the channel connection pool
	case c.pool <- cl:
	default:
		cl.Quit()
	}
}
