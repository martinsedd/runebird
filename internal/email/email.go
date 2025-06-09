package email

import (
	"fmt"
	"net/smtp"
	"runebird/internal/config"
)

type Sender struct {
	cfg  *config.SMTPConfig
	auth smtp.Auth
	from string
}

func New(cfg *config.SMTPConfig) (*Sender, error) {
	if cfg.Host == "" || cfg.Port == 0 || cfg.Username == "" || cfg.Password == "" || cfg.FromAddress == "" {
		return nil, fmt.Errorf("invalid SMTP configuration: missing required fields")
	}

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	return &Sender{
		cfg:  cfg,
		auth: auth,
		from: cfg.FromAddress,
	}, nil
}

func (s *Sender) Send(recipients []string, subject, htmlBody string) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients provided")
	}

	msg := []byte(fmt.Sprintf(
		"To: %s\r\n"+
			"From: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n",
		joinRecipients(recipients), s.from, subject, htmlBody))

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	err := smtp.SendMail(addr, s.auth, s.from, recipients, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}

func joinRecipients(recipients []string) string {
	if len(recipients) == 0 {
		return ""
	}
	result := recipients[0]
	for i := 1; i < len(recipients); i++ {
		result += ", " + recipients[i]
	}
	return result
}
