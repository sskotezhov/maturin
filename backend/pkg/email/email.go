package email

import (
	"fmt"
	"html"
	"strings"

	"gopkg.in/gomail.v2"
)

func (s *Sender) SendInquirySubmitted(to string, inquiryID uint, name, phone, comment string) error {
	body := buildInquiryEmailHTML(inquiryID, name, phone, comment)
	return s.send(to, fmt.Sprintf("Новая анкета #%d — Матурин", inquiryID), body)
}

func (s *Sender) SendOrderSubmitted(to string, orderID uint) error {
	body := buildOrderEmailHTML(
		"Новая заявка",
		fmt.Sprintf("Поступила новая заявка #%d. Войдите в систему для её рассмотрения.", orderID),
	)
	return s.send(to, fmt.Sprintf("Новая заявка #%d — Матурин", orderID), body)
}

func (s *Sender) SendOrderApproved(to string, orderID uint, totalPrice float64) error {
	body := buildOrderEmailHTML(
		"Заявка согласована",
		fmt.Sprintf("Ваша заявка #%d согласована. Итоговая сумма: %.2f ₽.", orderID, totalPrice),
	)
	return s.send(to, fmt.Sprintf("Заявка #%d согласована — Матурин", orderID), body)
}

func (s *Sender) SendNewMessage(to string, orderID uint) error {
	body := buildOrderEmailHTML(
		"Новое сообщение",
		fmt.Sprintf("В заявке #%d появилось новое сообщение. Войдите в систему для просмотра.", orderID),
	)
	return s.send(to, fmt.Sprintf("Новое сообщение в заявке #%d — Матурин", orderID), body)
}

func buildInquiryEmailHTML(inquiryID uint, name, phone, comment string) string {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		comment = "Комментарий не указан."
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Новая анкета</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f5;padding:40px 0;">
    <tr>
      <td align="center">
        <table width="520" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
          <tr>
            <td style="background-color:#6644aa;padding:28px 40px;text-align:center;">
              <span style="font-size:22px;font-weight:700;color:#ffffff;letter-spacing:0.5px;">МАТУРИН</span>
            </td>
          </tr>
          <tr>
            <td style="padding:40px 40px 32px;">
              <h2 style="margin:0 0 12px;font-size:20px;font-weight:600;color:#333333;">Новая анкета #%d</h2>
              <p style="margin:0 0 18px;font-size:15px;color:#6b7280;line-height:1.6;">Поступила новая заявка с формы обратной связи.</p>
              <table width="100%%" cellpadding="0" cellspacing="0" style="font-size:14px;color:#333333;line-height:1.6;">
                <tr><td style="font-weight:600;width:110px;padding:4px 0;">Имя:</td><td style="padding:4px 0;">%s</td></tr>
                <tr><td style="font-weight:600;width:110px;padding:4px 0;">Телефон:</td><td style="padding:4px 0;">%s</td></tr>
                <tr><td style="font-weight:600;width:110px;padding:4px 0;vertical-align:top;">Комментарий:</td><td style="padding:4px 0;">%s</td></tr>
              </table>
            </td>
          </tr>
          <tr>
            <td style="background-color:#f9fafb;border-top:1px solid #e5e7eb;padding:20px 40px;text-align:center;">
              <p style="margin:0;font-size:12px;color:#9ca3af;">© 2026 Матурин — программы и оборудование</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, inquiryID, html.EscapeString(name), html.EscapeString(phone), html.EscapeString(comment))
}

func buildOrderEmailHTML(title, description string) string {
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
          <tr>
            <td style="background-color:#6644aa;padding:28px 40px;text-align:center;">
              <span style="font-size:22px;font-weight:700;color:#ffffff;letter-spacing:0.5px;">МАТУРИН</span>
            </td>
          </tr>
          <tr>
            <td style="padding:40px 40px 32px;">
              <h2 style="margin:0 0 12px;font-size:20px;font-weight:600;color:#333333;">%s</h2>
              <p style="margin:0;font-size:15px;color:#6b7280;line-height:1.6;">%s</p>
            </td>
          </tr>
          <tr>
            <td style="background-color:#f9fafb;border-top:1px solid #e5e7eb;padding:20px 40px;text-align:center;">
              <p style="margin:0;font-size:12px;color:#9ca3af;">© 2026 Матурин — программы и оборудование</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, title, title, description)
}

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
	m.SetHeader("From", s.cfg.User)
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
              <p style="margin:0;font-size:12px;color:#9ca3af;">© 2026 Матурин — программы и оборудование</p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, title, title, description, code, disclaimer)
}
