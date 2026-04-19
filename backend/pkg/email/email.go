package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
}

type Sender struct {
	cfg Config
}

func NewSender(cfg Config) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) SendVerificationCode(to, code string) error {
	return s.send(to, "Подтверждение email", fmt.Sprintf(
		"Ваш код подтверждения: <b>%s</b><br>Код действителен 15 минут.", code,
	))
}

func (s *Sender) SendPasswordResetCode(to, code string) error {
	return s.send(to, "Восстановление пароля", fmt.Sprintf(
		"Ваш код для сброса пароля: <b>%s</b><br>Код действителен 15 минут.", code,
	))
}

func (s *Sender) send(to, subject, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.User)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(s.cfg.Host, s.cfg.Port, s.cfg.User, s.cfg.Password)
	return d.DialAndSend(m)
}
