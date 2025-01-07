package email

import (
	"fmt"
	"log"
	"net/smtp"
)

type EmailService interface {
	SendVerificationEmail(email, code string) error
	SendVerificationEmailAsync(email, code string)
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

func (s *SMTPEmailService) SendVerificationEmailAsync(email, code string) {
	go func() {
		err := s.SendVerificationEmail(email, code)
		if err != nil {
			fmt.Printf("Failed to send verification email to %s: %v\n", email, err)
		}
	}()
}

func (s *SMTPEmailService) SendVerificationEmail(email, code string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.smtpHost)
	htmlMessage := GenerateVerificationEmail(code)

	message := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: Email Verification\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
			"%s",
		s.from, email, htmlMessage,
	))
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	err := smtp.SendMail(addr, auth, s.from, []string{email}, message)
	if err != nil {
		return fmt.Errorf("Failed to send verification email: %w", err)
	}
	log.Printf("Sent verification emai to %s with code %s", email, code)
	return nil
}
