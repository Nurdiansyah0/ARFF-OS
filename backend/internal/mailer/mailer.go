package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	FromName string
	FromEmail string
}

func LoadConfig() *Config {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}

	fromName := os.Getenv("SMTP_FROM_NAME")
	if fromName == "" {
		fromName = "ARFF-OS Airport Security"
	}

	fromEmail := os.Getenv("SMTP_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = os.Getenv("SMTP_USER")
		if fromEmail == "" {
			fromEmail = "no-reply@arff-os.com"
		}
	}

	return &Config{
		Host:      host,
		Port:      port,
		Username:  os.Getenv("SMTP_USER"),
		Password:  os.Getenv("SMTP_PASSWORD"),
		FromName:  fromName,
		FromEmail: fromEmail,
	}
}

type Mailer struct {
	cfg *Config
}

func NewMailer() *Mailer {
	return &Mailer{cfg: LoadConfig()}
}

// SendOTPEmail sends an HTML 6-digit OTP verification email
func (m *Mailer) SendOTPEmail(toEmail, toName, otpCode string) error {
	subject := fmt.Sprintf("Kode Verifikasi ARFF-OS: %s", otpCode)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; margin: 0; padding: 20px; }
    .container { max-width: 520px; margin: 0 auto; background: #ffffff; border-radius: 12px; border: 1px solid #e2e8f0; overflow: hidden; }
    .header { background: #C62828; padding: 24px; text-align: center; color: #ffffff; }
    .header h1 { margin: 0; font-size: 20px; letter-spacing: 1px; }
    .content { padding: 32px 24px; text-align: center; color: #1e293b; }
    .otp-box { background: #f1f5f9; border: 2px dashed #cbd5e1; border-radius: 8px; font-size: 32px; font-weight: 800; letter-spacing: 6px; color: #C62828; padding: 18px; margin: 24px 0; font-family: monospace; }
    .footer { background: #f8fafc; border-top: 1px solid #e2e8f0; padding: 16px; text-align: center; font-size: 12px; color: #64748b; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>ARFF-OS SECURITY</h1>
      <p style="margin: 4px 0 0; font-size: 13px; opacity: 0.9;">Airport Rescue & Firefighting Operating System</p>
    </div>
    <div class="content">
      <h2 style="font-size: 18px; margin-top: 0;">Verifikasi Akun Administrator Bandara</h2>
      <p style="font-size: 14px; line-height: 1.5; color: #475569;">
        Halo <strong>%s</strong>, gunakan kode One-Time Password (OTP) berikut untuk menyelesaikan pendaftaran bandara Anda di platform ARFF-OS:
      </p>
      <div class="otp-box">%s</div>
      <p style="font-size: 13px; color: #64748b; margin-bottom: 0;">
        Kode ini hanya berlaku selama <strong>10 Menit</strong>. Jangan berikan kode ini kepada siapa pun.
      </p>
    </div>
    <div class="footer">
      &copy; ARFF-OS Platform &bull; Automated Security Mailer
    </div>
  </div>
</body>
</html>`, toName, otpCode)

	// If SMTP is not yet configured, log clearly in terminal for local testing
	if m.cfg.Username == "" || m.cfg.Password == "" {
		log.Printf("\n=======================================================\n"+
			"[MAILER DEV NOTICE] SMTP credentials not set.\n"+
			"Recipient: %s (%s)\n"+
			"OTP CODE : [ %s ] (Valid for 10 minutes)\n"+
			"=======================================================\n",
			toEmail, toName, otpCode)
		return nil
	}

	// Send via SMTP (Gmail / Brevo / Resend)
	return m.sendSMTP(toEmail, subject, body)
}

func (m *Mailer) sendSMTP(to, subject, htmlBody string) error {
	addr := net.JoinHostPort(m.cfg.Host, m.cfg.Port)

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", m.cfg.FromName, m.cfg.FromEmail)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	// Connect to server
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server: %w", err)
	}
	defer client.Close()

	// Upgrade to TLS if using port 587
	tlsConfig := &tls.Config{
		ServerName: m.cfg.Host,
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender and recipient
	if err := client.Mail(m.cfg.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	log.Printf("[MAILER SUCCESS] OTP email successfully dispatched to %s via %s", to, m.cfg.Host)
	return client.Quit()
}
