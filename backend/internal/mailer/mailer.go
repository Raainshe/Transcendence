package mailer

import (
	"fmt"
	"net/smtp"
)

type Mailer interface {
	Send(to, subject, body string) error
}

type SMTPMailer struct {
	host string
	port string
	username string
	password string
	from string
}

func NewSMTPMailer(host string, port string, username string, password string, from string) *SMTPMailer {
	return &SMTPMailer{host: host, port: port, username: username, password: password, from: from}
}

func (m *SMTPMailer) Send(to string, subject string, body string) error {
	addr := m.host + ":" + m.port
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	msg := fmt.Appendf(nil, `From: %s\r\n
					   		 To: %s\r\n
					   		 Subject: %s\r\n
					   		 MIME-Version: 1.0\r\n
							 Content-Type: text/plain; charset=UTF-8\r\n\r\n
					   		 %s\r\n`,
							 m.from, to, subject, body)
	return smtp.SendMail(addr, auth, m.from, []string{to}, msg)
}
