package wa

import (
	"fmt"
	"github.com/JMURv/sso/internal/config"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"go.uber.org/zap"
	"net/http"
)

type Port interface {
	BeginLogin(user webauthn.User, opts ...webauthn.LoginOption) (*protocol.CredentialAssertion, *webauthn.SessionData, error)
	FinishLogin(user webauthn.User, session webauthn.SessionData, response *http.Request) (*webauthn.Credential, error)
	BeginRegistration(user webauthn.User, opts ...webauthn.RegistrationOption) (creation *protocol.CredentialCreation, session *webauthn.SessionData, err error)
	FinishRegistration(user webauthn.User, session webauthn.SessionData, response *http.Request) (*webauthn.Credential, error)
}

type SessionType string

const (
	Login    SessionType = "login"
	Register SessionType = "registration"
)

type WAuthn struct {
	*webauthn.WebAuthn
}

func New(conf config.Config) *WAuthn {
	origins := make([]string, 0, len(conf.Auth.WebAuthn.Origins)+1)
	origins = append(origins, fmt.Sprintf("%v://%v", conf.Server.Scheme, conf.Server.Domain))
	origins = append(origins, conf.Auth.WebAuthn.Origins...)
	wa, err := webauthn.New(
		&webauthn.Config{
			RPDisplayName: conf.ServiceName,
			RPID:          conf.Server.Domain,
			RPOrigins:     origins,
		},
	)
	if err != nil {
		zap.L().Fatal("Failed to initialize WebAuthn", zap.Error(err))
	}

	return &WAuthn{wa}
}
