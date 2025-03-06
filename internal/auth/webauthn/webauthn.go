package wa

import (
	"fmt"
	"github.com/JMURv/sso/internal/config"
	"github.com/go-webauthn/webauthn/webauthn"
	"go.uber.org/zap"
)

type WAuthn struct {
	wa *webauthn.WebAuthn
}

func New(conf config.Config) *WAuthn {
	wconf := &webauthn.Config{
		RPDisplayName: conf.ServiceName,
		RPID:          conf.Server.Domain,
		RPOrigins: []string{
			fmt.Sprintf("%v://%v", conf.Server.Scheme, conf.Server.Domain),
		},
	}

	wa, err := webauthn.New(wconf)
	if err != nil {
		zap.L().Fatal("Failed to initialize WebAuthn", zap.Error(err))
	}

	return &WAuthn{
		wa: wa,
	}
}
