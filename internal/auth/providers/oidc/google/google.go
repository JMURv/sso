package providers

import (
	providers "github.com/JMURv/sso/internal/auth/providers/oidc"
	"github.com/JMURv/sso/internal/config"
)

type GoogleOIDC struct {
	*providers.OIDCProvider
}

func NewGoogleOIDC(conf config.Config) *GoogleOIDC {
	provider, err := providers.NewOIDCProvider(
		"https://accounts.google.com",
		conf.Auth.OIDC.Google.ClientID,
		conf.Auth.OIDC.Google.ClientSecret,
		conf.Auth.OIDC.Google.RedirectURL,
		conf.Auth.OIDC.Google.Scopes,
		conf.Auth.OIDC.SuccessURL,
	)

	if err != nil {
		return nil
	}

	return &GoogleOIDC{provider}
}
