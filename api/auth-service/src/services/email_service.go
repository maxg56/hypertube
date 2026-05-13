package services

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

type EmailService struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

func NewEmailService() *EmailService {
	return &EmailService{
		SMTPHost:     getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvOrDefault("SMTP_PORT", "587"),
		SMTPUsername: getEnvOrDefault("SMTP_USERNAME", ""),
		SMTPPassword: getEnvOrDefault("SMTP_PASSWORD", ""),
		FromEmail:    getEnvOrDefault("FROM_EMAIL", "noreply@hypertube.app"),
		FromName:     getEnvOrDefault("FROM_NAME", "Hypertube"),
	}
}

type VerificationEmailData struct {
	VerificationCode string
}

func (es *EmailService) SendVerificationEmail(toEmail, verificationCode string) error {
	if es.SMTPUsername == "" || es.SMTPPassword == "" {
		log.Printf("email verification code for %s: %s (SMTP not configured)", toEmail, verificationCode)
		return nil
	}

	tmpl, err := template.ParseFiles(filepath.Join("templates", "email", "verification_email.html"))
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body strings.Builder
	if err := tmpl.Execute(&body, VerificationEmailData{VerificationCode: verificationCode}); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	return es.sendEmail(toEmail, "Vérification de votre email - Hypertube", body.String())
}

type PasswordResetData struct {
	ResetURL string
}

func (es *EmailService) SendPasswordResetEmail(toEmail, resetToken string) error {
	if es.SMTPUsername == "" || es.SMTPPassword == "" {
		resetURL := fmt.Sprintf("%s/reinitialiser-mot-de-passe?token=%s", getEnvOrDefault("FRONTEND_URL", "https://localhost:8443"), resetToken)
		log.Printf("password reset link for %s: %s (SMTP not configured)", toEmail, resetURL)
		return nil
	}

	frontendURL := getEnvOrDefault("FRONTEND_URL", "https://localhost:8443")
	resetURL := fmt.Sprintf("%s/reinitialiser-mot-de-passe?token=%s", frontendURL, resetToken)

	tmpl, err := template.ParseFiles(filepath.Join("templates", "email", "password_reset.html"))
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body strings.Builder
	if err := tmpl.Execute(&body, PasswordResetData{ResetURL: resetURL}); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	return es.sendEmail(toEmail, "Réinitialisation de votre mot de passe - Hypertube", body.String())
}

func (es *EmailService) sendEmail(to, subject, body string) error {
	port, err := strconv.Atoi(es.SMTPPort)
	if err != nil {
		return fmt.Errorf("invalid SMTP port %q: %v", es.SMTPPort, err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(es.FromEmail, es.FromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(es.SMTPHost, port, es.SMTPUsername, es.SMTPPassword)
	if port == 465 {
		d.SSL = true
	}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
