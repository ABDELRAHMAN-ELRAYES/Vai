package mailer

import "time"

type VerifyEmailData struct {
	Name      string
	VerifyURL string
	SentAt    time.Time
}

type WelcomeEmailData struct {
	Name         string
	DashboardURL string
	SentAt       time.Time
}

type ResetPasswordData struct {
	Name       string
	ResetURL   string
	Expiration string
}
