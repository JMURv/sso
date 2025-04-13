package providers

import (
	"context"
	providers "github.com/JMURv/sso/internal/auth/providers/oauth2"
	conf "github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/goccy/go-json"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"io"
	"net/http"
)

type GoogleProvider struct {
	*providers.OAuth2Provider
}

func NewGoogleOAuth2(config conf.Config) *GoogleProvider {
	provider := providers.NewOAuth2Provider(
		config.Auth.Oauth.Google.ClientID,
		config.Auth.Oauth.Google.ClientSecret,
		config.Auth.Oauth.Google.RedirectURL,
		google.Endpoint,
		config.Auth.Oauth.Google.Scopes,
		config.Auth.Oauth.SuccessURL,
		func(ctx context.Context, token string, cli *http.Client) (*dto.ProviderResponse, error) {
			const op = "provider.GetUser.google"
			span, ctx := opentracing.StartSpanFromContext(ctx, op)
			defer span.Finish()

			resp, err := cli.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token)
			if err != nil {
				return nil, err
			}
			defer func(Body io.ReadCloser) {
				if err := Body.Close(); err != nil {
					zap.L().Debug("failed to close body", zap.Error(err))
				}
			}(resp.Body)

			var gRes dto.GoogleResponse
			if err = json.NewDecoder(resp.Body).Decode(&gRes); err != nil {
				return nil, err
			}

			return &dto.ProviderResponse{
				ProviderID: gRes.ID,
				Email:      gRes.Email,
				Name:       gRes.Name,
			}, nil
		},
	)
	return &GoogleProvider{provider}
}
