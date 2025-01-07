package email

import (
	"fmt"
	"log"
	"net/smtp"
)

type EmailService interface {
	SendVerificationEmail(email, code string) error
}

type SMTPEmailService struct {
	smtpHost string
	smtpPort string
	username string
	password string
	from     string
}

func NewSMTPEmailService(host, port, username, password, from string) *SMTPEmailService {
	return &SMTPEmailService{
		smtpHost: host,
		smtpPort: port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SMTPEmailService) SendVerificationEmail(email, code string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.smtpHost)
	message := []byte(fmt.Sprintf(
		"Subject: Email Verification\n\nYour verification code is %s\n", code,
	))
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	err := smtp.SendMail(addr, auth, s.from, []string{email}, message)
	if err != nil {
		return fmt.Errorf("Failed to send verification email: %w", err)
	}
	log.Printf("Sent verification emai to %s with code %s", email, code)
	return nil
}
