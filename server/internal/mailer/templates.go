package mailer

import (
	"bytes"
	"fmt"
	"strings"
)

// execute: render template - fill the email template (two chunks: subject & body ) with the required data
func (c *client) execute(templateFile string, data any) (subject, body string, err error) {
	var subjectBuf, bodyBuf bytes.Buffer

	if err := c.tmpl.ExecuteTemplate(&subjectBuf, templateFile+"_subject", data); err != nil {
		return "", "", fmt.Errorf("execute subject in %q: %w", templateFile, err)
	}
	if err := c.tmpl.ExecuteTemplate(&bodyBuf, templateFile+"_body", data); err != nil {
		return "", "", fmt.Errorf("execute body in %q: %w", templateFile, err)
	}

	return strings.TrimSpace(subjectBuf.String()), bodyBuf.String(), nil
}
