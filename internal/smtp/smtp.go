package smtp

import (
	"context"
	"fmt"
	"github.com/JMURv/sso/internal/config"
	md "github.com/JMURv/sso/internal/models"
	"gopkg.in/gomail.v2"
)

type EmailServer struct {
	server       string
	port         int
	user         string
	pass         string
	admin        string
	serverConfig *config.ServerConfig
}

func New(conf *config.EmailConfig, serverConfig *config.ServerConfig) *EmailServer {
	return &EmailServer{
		server:       conf.Server,
		port:         conf.Port,
		user:         conf.User,
		pass:         conf.Pass,
		admin:        conf.Admin,
		serverConfig: serverConfig,
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
		return err
	}
	return nil
}

func (s *EmailServer) SendLoginEmail(_ context.Context, code int, toEmail string) error {
	m := s.GetMessageBase("Login Code", toEmail)
	m.SetBody("text/plain", fmt.Sprintf("Login code: %v", code))
	return s.Send(m)
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
