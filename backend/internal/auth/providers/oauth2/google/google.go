package google

import (
	"context"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/goccy/go-json"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
)

type Provider struct {
	config *oauth2.Config
}

func New(conf config.Config) *Provider {
	return &Provider{
		config: &oauth2.Config{
			ClientID:     conf.Auth.Providers.Oauth.Google.ClientID,
			ClientSecret: conf.Auth.Providers.Oauth.Google.ClientSecret,
			RedirectURL:  conf.Auth.Providers.Oauth.Google.RedirectURL,
			Scopes:       conf.Auth.Providers.Oauth.Google.Scopes,
			Endpoint:     google.Endpoint,
		},
	}
}

func (p *Provider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *Provider) Exchange(ctx context.Context, code string) (*dto.ProviderResponse, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		zap.L().Error("failed to exchange token", zap.Error(err))
		return nil, err
	}

	return p.getUserInfo(ctx, token)
}

func (p *Provider) getUserInfo(ctx context.Context, token *oauth2.Token) (*dto.ProviderResponse, error) {
	const op = "provider.GetUser.google"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	cli := p.config.Client(ctx, token)
	resp, err := cli.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			zap.L().Error("failed to close body", zap.Error(err))
		}
	}(resp.Body)

	var gRes dto.GoogleResponse
	if err = json.NewDecoder(resp.Body).Decode(&gRes); err != nil {
		zap.L().Error("failed to decode body", zap.Error(err))
		return nil, err
	}

	return &dto.ProviderResponse{
		ProviderID: gRes.ID,
		Email:      gRes.Email,
		Name:       gRes.Name,
	}, nil
}
