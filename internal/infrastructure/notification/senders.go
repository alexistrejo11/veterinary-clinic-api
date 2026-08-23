package notification

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"clinic-vet-api/internal/core/notifications"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type SMTPSender struct {
	config SMTPConfig
}

type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
}

func NewSMTPSender(emailConfig SMTPConfig) (*SMTPSender, error) {
	if emailConfig.Host == "" || emailConfig.Port == 0 || emailConfig.FromEmail == "" {
		return nil, errors.New("SMTP host, port, and sender email are required")
	}
	return &SMTPSender{config: emailConfig}, nil
}

var _ notifications.EmailSender = (*SMTPSender)(nil)

func (s *SMTPSender) Send(ctx context.Context, notification *notifications.Notification) (string, error) {
	if notification == nil || notification.UserEmail == "" {
		return "", errors.New("notification recipient email is required")
	}
	if containsHeaderBreak(notification.UserEmail) || containsHeaderBreak(notification.Subject) {
		return "", errors.New("notification contains invalid header characters")
	}

	address := net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port))
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return "", fmt.Errorf("connect to SMTP server: %w", err)
	}

	if s.config.Port == 465 {
		conn = tls.Client(conn, &tls.Config{ServerName: s.config.Host, MinVersion: tls.VersionTLS12})
	}
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		_ = conn.Close()
		return "", fmt.Errorf("create SMTP client: %w", err)
	}
	defer client.Close()

	if s.config.Port != 465 {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return "", errors.New("SMTP server does not support STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{ServerName: s.config.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return "", fmt.Errorf("start SMTP TLS: %w", err)
		}
	}
	if s.config.Username != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return "", fmt.Errorf("authenticate SMTP client: %w", err)
		}
	}
	if err := client.Mail(s.config.FromEmail); err != nil {
		return "", fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(notification.UserEmail); err != nil {
		return "", fmt.Errorf("set SMTP recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return "", fmt.Errorf("open SMTP message: %w", err)
	}
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.config.FromEmail, notification.UserEmail, notification.Subject, notification.Message)
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return "", fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("send SMTP message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return "", err
	}
	return "smtp", nil
}

type TwilioSender struct {
	client *twilio.RestClient
	from   string
}

func NewTwilioSender(client *twilio.RestClient, from string) (*TwilioSender, error) {
	if client == nil || from == "" {
		return nil, errors.New("Twilio client and sender phone are required")
	}
	return &TwilioSender{client: client, from: from}, nil
}

var _ notifications.SMSender = (*TwilioSender)(nil)

func (s *TwilioSender) Send(ctx context.Context, notification *notifications.Notification) (string, error) {
	if notification == nil || notification.UserPhone == "" {
		return "", errors.New("notification recipient phone is required")
	}
	params := &openapi.CreateMessageParams{}
	params.SetTo(notification.UserPhone)
	params.SetFrom(s.from)
	params.SetBody(notification.Message)
	message, err := s.client.Api.CreateMessage(params)
	if err != nil {
		return "", fmt.Errorf("send Twilio SMS: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if message.Sid == nil {
		return "", errors.New("Twilio returned no message reference")
	}
	return *message.Sid, nil
}

func containsHeaderBreak(value string) bool {
	return strings.ContainsAny(value, "\r\n")
}
