package smtp

import (
	"context"
	"fmt"
	"github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type EmailServer struct {
	server       string
	port         int
	user         string
	pass         string
	admin        string
	serverConfig config.ServerConfig
}

func New(conf config.Config) *EmailServer {
	return &EmailServer{
		server:       conf.Email.Server,
		port:         conf.Email.Port,
		user:         conf.Email.User,
		pass:         conf.Email.Pass,
		admin:        conf.Email.Admin,
		serverConfig: conf.Server,
	}
}

func (s *EmailServer) GetMessageBase(subject string, toEmail string) *gomail.Message {
	m := gomail.NewMessage()
	m.SetHeader("From", s.user)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	return m
}

func (s *EmailServer) Send(m *gomail.Message) error {
	d := gomail.NewDialer(s.server, s.port, s.user, s.pass)
	if err := d.DialAndSend(m); err != nil {
		zap.L().Error(
			"Failed to send an email",
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (s *EmailServer) SendLoginEmail(_ context.Context, code int, toEmail string) {
	m := s.GetMessageBase("Login Code", toEmail)
	m.SetBody("text/plain", fmt.Sprintf("Login code: %v", code))
	_ = s.Send(m)
}

func (s *EmailServer) SendActivationCodeEmail(_ context.Context, code uint64, toEmail string) error {
	m := s.GetMessageBase("Activation Code", toEmail)
	m.SetBody("text/plain", fmt.Sprintf("Activation code: %v", code))
	return s.Send(m)
}

func (s *EmailServer) SendForgotPasswordEmail(_ context.Context, token, uid64, toEmail string) error {
	m := s.GetMessageBase("Forgot Password Code", toEmail)

	params := fmt.Sprintf("?uidb64=%v&token=%v", uid64, token)
	resetURL := fmt.Sprintf("%v://%v/email/password-reset/%v", s.serverConfig.Scheme, s.serverConfig.Domain, params)

	m.SetBody("text/plain", fmt.Sprintf("Forgot password URL: %v", resetURL))
	return s.Send(m)
}

func (s *EmailServer) SendSupportEmail(_ context.Context, u *md.User, theme, text string) error {
	m := s.GetMessageBase(theme, s.admin)
	m.SetBody("text/plain", fmt.Sprintf("New support message from %v with email: %v\n %v", u.Name, u.Email, text))
	return s.Send(m)
}

func (s *EmailServer) SendUserCredentials(_ context.Context, email, pass string) error {
	m := s.GetMessageBase("Данные для входа", email)
	m.SetBody("text/plain", fmt.Sprintf("Данные для входа.\nLogin: %v\nPassword: %v", email, pass))
	return s.Send(m)
}
