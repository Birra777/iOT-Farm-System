package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Mailer sends plain-text emails via SMTP with STARTTLS (port 587)
// or implicit TLS (port 465). If SMTPHost is empty it is a no-op.
type Mailer struct {
	host string
	port string
	user string
	pass string
	from string
}

// New returns a configured Mailer. Pass an empty host to disable email.
func New(host, port, user, pass string) *Mailer {
	return &Mailer{host: host, port: port, user: user, pass: pass, from: user}
}

// Enabled reports whether the mailer is configured to send emails.
func (m *Mailer) Enabled() bool {
	return m.host != ""
}

// Send delivers a plain-text email to one or more comma-separated recipients.
func (m *Mailer) Send(to, subject, body string) error {
	if !m.Enabled() {
		return nil
	}

	recipients := strings.Split(to, ",")
	for i, r := range recipients {
		recipients[i] = strings.TrimSpace(r)
	}

	msg := buildMessage(m.from, to, subject, body)

	addr := net.JoinHostPort(m.host, m.port)

	// Port 465 → implicit TLS (SMTPS). Port 587 (and everything else) → STARTTLS.
	if m.port == "465" {
		return sendSMTPS(addr, m.host, m.user, m.pass, m.from, recipients, msg)
	}
	return sendSTARTTLS(addr, m.host, m.user, m.pass, m.from, recipients, msg)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: AgriStream Alerts <" + from + ">\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

func sendSTARTTLS(addr, host, user, pass, from string, to []string, msg []byte) error {
	auth := smtp.PlainAuth("", user, pass, host)
	return smtp.SendMail(addr, auth, from, to, msg)
}

func sendSMTPS(addr, host, user, pass, from string, to []string, msg []byte) error {
	tlsCfg := &tls.Config{ServerName: host}
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", user, pass, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	for _, r := range to {
		if err := client.Rcpt(r); err != nil {
			return fmt.Errorf("smtp RCPT TO %s: %w", r, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	return w.Close()
}
