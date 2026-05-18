package provider

import (
	"fmt"
	"log"
	"net/smtp"
)

// SMTPProvider sends emails via a real SMTP server.
type SMTPProvider struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPProvider(host, port, username, password, from string) *SMTPProvider {
	return &SMTPProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SMTPProvider) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.from, to, subject, body)

	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	log.Printf("[SMTPProvider] Email sent to=%s subject=%q", to, subject)
	return nil
}
