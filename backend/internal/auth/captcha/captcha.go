package captcha

import (
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
)

type Port interface {
	VerifyRecaptcha(token string, action Actions) (bool, error)
}

type Actions string

const (
	PassAuth   Actions = "pass_auth"
	EmailAuth  Actions = "email_auth"
	ForgotPass Actions = "forgot_pass"
	WALogin    Actions = "wa_login"
)

type Core struct {
	secret string
}

func New(conf config.Config) *Core {
	return &Core{
		secret: conf.Auth.Captcha.Secret,
	}
}

func (c *Core) VerifyRecaptcha(token string, action Actions) (bool, error) {
	resp, err := http.PostForm(
		"https://www.google.com/recaptcha/api/siteverify",
		url.Values{
			"secret":   {c.secret},
			"response": {token},
		},
	)
	if err != nil {
		zap.L().Error("failed to verify recaptcha", zap.Error(err))
		return false, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			zap.L().Error("failed to close body", zap.Error(err))
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("failed to read body", zap.Error(err))
		return false, err
	}

	var result dto.RecaptchaResponse
	if err = json.Unmarshal(body, &result); err != nil {
		zap.L().Error("failed to unmarshal body", zap.Error(err))
		return false, err
	}

	return result.Success && result.Score > 0.5 && result.Action == string(action), nil
}
