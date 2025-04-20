package wa

import (
	"fmt"
	"github.com/JMURv/sso/internal/config"
	"github.com/go-webauthn/webauthn/webauthn"
	"go.uber.org/zap"
)

type SessionType string

const (
	Login    SessionType = "login"
	Register SessionType = "registration"
)

type WAuthn struct {
	*webauthn.WebAuthn
}

func New(conf config.Config) *WAuthn {
	wa, err := webauthn.New(
		&webauthn.Config{
			RPDisplayName: conf.ServiceName,
			RPID:          conf.Server.Domain,
			RPOrigins: []string{
				fmt.Sprintf("%v://%v", conf.Server.Scheme, conf.Server.Domain),
				"http://localhost",
				"http://127.0.0.1",
			},
		},
	)
	if err != nil {
		zap.L().Fatal("Failed to initialize WebAuthn", zap.Error(err))
	}

	return &WAuthn{wa}
}
