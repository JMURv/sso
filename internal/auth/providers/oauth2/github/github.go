package providers

import (
	"context"
	providers "github.com/JMURv/sso/internal/auth/providers/oauth2"
	conf "github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/goccy/go-json"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type GitHubProvider struct {
	*providers.OAuth2Provider
}

type googleResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Picture    string `json:"picture"`
	VerifEmail bool   `json:"verified_email"`
}

// TODO: refactor
func NewGitHubOAuth2(config conf.Config) *GitHubProvider {
	provider := providers.NewOAuth2Provider(
		config.Auth.Oauth.GitHub.ClientID,
		config.Auth.Oauth.GitHub.ClientSecret,
		config.Auth.Oauth.GitHub.RedirectURL,
		github.Endpoint,
		config.Auth.Oauth.GitHub.Scopes,
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

			var gRes googleResponse
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
