package email

import (
	"fmt"
	"log"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/internal/models"
	"gopkg.in/mail.v2"
)

// EmailOptions represents options for sending an email
type EmailOptions struct {
	To      []string
	Subject string
	HTML    string
	Text    string
}

// SignupEmailData represents data for signup email
type SignupEmailData struct {
	UserEmail  string
	UserName   string
	SignupLink string
}

// Client represents an email client
type Client struct {
	dialer *mail.Dialer
	config *config.Config
}

// NewClient creates a new email client
func NewClient(cfg *config.Config) *Client {
	d := mail.NewDialer(cfg.EmailHost, cfg.EmailPort, cfg.EmailUser, cfg.EmailPassword)

	// Set secure based on port (465 = SSL, others = STARTTLS)
	if cfg.EmailPort == 465 {
		d.SSL = true
	} else {
		d.StartTLSPolicy = mail.MandatoryStartTLS
	}

	return &Client{
		dialer: d,
		config: cfg,
	}
}

// SendEmail sends a generic email
func (c *Client) SendEmail(options EmailOptions) error {
	m := mail.NewMessage()

	// Set sender
	m.SetHeader("From", fmt.Sprintf("%s <%s>", c.config.EmailFromName, c.config.EmailFrom))

	// Set recipients
	m.SetHeader("To", options.To...)

	// Set subject
	m.SetHeader("Subject", options.Subject)

	// Set body
	m.SetBody("text/html", options.HTML)
	if options.Text != "" {
		m.AddAlternative("text/plain", options.Text)
	} else {
		m.AddAlternative("text/plain", options.Subject)
	}

	// Send email
	if err := c.dialer.DialAndSend(m); err != nil {
		log.Printf("Error sending email: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully to: %v", options.To)
	return nil
}

// SendNotificationEmail sends a notification email
func (c *Client) SendNotificationEmail(notification *models.Notification, userEmail string) error {
	html := getNotificationEmailTemplate(notification)

	return c.SendEmail(EmailOptions{
		To:      []string{userEmail},
		Subject: notification.Title,
		HTML:    html,
		Text:    notification.Message,
	})
}

// SendAlertEmail sends an alert email
func (c *Client) SendAlertEmail(to []string, title, message string, severity string) error {
	if severity == "" {
		severity = "info"
	}

	html := getAlertEmailTemplate(title, message, severity)

	return c.SendEmail(EmailOptions{
		To:      to,
		Subject: title,
		HTML:    html,
		Text:    message,
	})
}

// SendSignupLinkEmail sends a signup link email
func (c *Client) SendSignupLinkEmail(data SignupEmailData) error {
	html := getSignupEmailTemplate(data)

	return c.SendEmail(EmailOptions{
		To:      []string{data.UserEmail},
		Subject: "Welcome! Complete Your Registration",
		HTML:    html,
		Text:    fmt.Sprintf("Hello %s, you have been invited to join our platform. Click here to complete your registration: %s", data.UserName, data.SignupLink),
	})
}

// VerifyConnection verifies the email transporter configuration
func (c *Client) VerifyConnection() error {
	closer, err := c.dialer.Dial()
	if err != nil {
		log.Printf("Email server connection error: %v", err)
		return fmt.Errorf("failed to connect to email server: %w", err)
	}
	defer closer.Close()

	log.Println("Email server is ready to send emails")
	return nil
}
