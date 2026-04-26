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
	body := buildEmailHTML(
		"Подтверждение email",
		"Для завершения регистрации введите код подтверждения:",
		code,
		"Код действителен 15 минут. Если вы не регистрировались на нашем сайте — просто проигнорируйте это письмо.",
	)
	return s.send(to, "Подтверждение email — Матурин", body)
}

func (s *Sender) SendPasswordResetCode(to, code string) error {
	body := buildEmailHTML(
		"Сброс пароля",
		"Для сброса пароля введите код подтверждения:",
		code,
		"Код действителен 15 минут. Если вы не запрашивали сброс пароля — просто проигнорируйте это письмо.",
	)
	return s.send(to, "Сброс пароля — Матурин", body)
}

func (s *Sender) send(to, subject, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("Матурин <%s>", s.cfg.User))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(s.cfg.Host, s.cfg.Port, s.cfg.User, s.cfg.Password)
	return d.DialAndSend(m)
}

func buildEmailHTML(title, description, code, disclaimer string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f5;padding:40px 0;">
    <tr>
      <td align="center">
        <table width="520" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">

          <!-- Header -->
          <tr>
            <td style="background-color:#6644aa;padding:28px 40px;text-align:center;">
              <span style="font-size:22px;font-weight:700;color:#ffffff;letter-spacing:0.5px;">МАТУРИН</span>
            </td>
          </tr>

          <!-- Body -->
          <tr>
            <td style="padding:40px 40px 32px;">
              <h2 style="margin:0 0 12px;font-size:20px;font-weight:600;color:#333333;">%s</h2>
              <p style="margin:0 0 28px;font-size:15px;color:#6b7280;line-height:1.6;">%s</p>

              <!-- Code block -->
              <div style="background-color:#f9fafb;border:2px solid #6644aa;border-radius:10px;padding:24px;text-align:center;margin-bottom:28px;">
                <span style="font-size:38px;font-weight:700;letter-spacing:10px;color:#6644aa;font-family:'Courier New',Courier,monospace;">%s</span>
              </div>

              <p style="margin:0;font-size:13px;color:#9ca3af;line-height:1.6;">%s</p>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="background-color:#f9fafb;border-top:1px solid #e5e7eb;padding:20px 40px;text-align:center;">
              <p style="margin:0;font-size:12px;color:#9ca3af;">© 2025 Матурин — программы и оборудование</p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, title, title, description, code, disclaimer)
}
