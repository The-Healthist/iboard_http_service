package base_services

import (
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"gopkg.in/gomail.v2"
)

type IEmailService interface {
	SendEmail(to []string, subject string, body string) error
}

type EmailService struct {
	emailClient *gomail.Dialer
}

func NewEmailService(emailClient *gomail.Dialer) IEmailService {
	log.Info("初始化邮件服务")
	return &EmailService{
		emailClient: emailClient,
	}
}

func (s *EmailService) SendEmail(to []string, subject string, body string) error {
	log.Info("尝试发送邮件 | 收件人: %v | 主题: %s", to, subject)

	m := gomail.NewMessage()
	m.SetHeader("From", s.emailClient.Username)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if err := s.emailClient.DialAndSend(m); err != nil {
		log.Error("发送邮件失败 | 收件人: %v | 主题: %s | 错误: %v", to, subject, err)
		return err
	}

	log.Info("发送邮件成功 | 收件人: %v | 主题: %s", to, subject)
	return nil
}
