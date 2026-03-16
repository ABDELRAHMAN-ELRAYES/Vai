package mailer

import (
	"fmt"
	"html/template"
	"net/smtp"
	"time"

	"github.com/ABDELRAHMAN-ELRAYES/Vai/internal/config"
)

type client struct {
	cfg  *config.Mail
	tmpl *template.Template // compiled templates set (all parsed templates which are in templates directory)

	pool chan *smtp.Client // SMTP connection pool - to reuse the connection instead of  creating a new connection with each email
}

// New: Create a New mailer client
func New(cfg *config.Mail) (Client, error) {

	// Allow Custom functions inside templates {{ formatDate .CreatedAt }}
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string { return t.Format("January 2, 2006") },
		"formatTime": func(t time.Time) string { return t.Format("3:04 PM MST") },
	}

	// Parse templates by:
	// 1. Create New template container
	// 2. Register Custom functions (so templates can call them)
	// 3. Load files from embedded file system

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(FS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("mailer: parse templates: %w", err)
	}

	return &client{
		cfg:  cfg,
		tmpl: tmpl,
		pool: make(chan *smtp.Client, 5), // pool size
	}, nil
}