package service

import (
	"github.com/kiry163/claw-pliers/internal/config"
	"github.com/kiry163/claw-pliers/internal/logger"
	"github.com/kiry163/claw-pliers/internal/mail"

	"github.com/rs/zerolog"
)

type MailService struct {
	logger *zerolog.Logger
}

func NewMailService() *MailService {
	l := logger.Get()
	return &MailService{
		logger: l,
	}
}

func (s *MailService) TestConnection(email string) (int64, error) {
	latency, err := mail.TestConnection(email)
	if err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("failed to test mail connection")
		return 0, err
	}

	s.logger.Info().Str("email", email).Int64("latency", latency).Msg("mail connection test successful")
	return latency, nil
}

func (s *MailService) SendMail(from, to, subject, body string) error {
	err := mail.SendMail(from, to, subject, body)
	if err != nil {
		s.logger.Error().Err(err).Str("from", from).Str("to", to).Msg("failed to send mail")
		return err
	}

	s.logger.Info().Str("from", from).Str("to", to).Str("subject", subject).Msg("mail sent successfully")
	return nil
}

func (s *MailService) GetLatestEmails(email string, count int) ([]mail.EmailSummary, error) {
	emails, err := mail.GetLatestEmails(email, count)
	if err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("failed to get latest emails")
		return nil, err
	}

	s.logger.Info().Str("email", email).Int("count", len(emails)).Msg("retrieved latest emails")
	return emails, nil
}

func (s *MailService) ListAccounts() []config.AccountConfig {
	return mail.ListAccounts()
}
