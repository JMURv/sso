package providers

import (
	"context"
	"fmt"
	"github.com/JMURv/sso/internal/dto"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/oauth2"
	"net/http"
)

type OAuth2Provider struct {
	successURL string
	config     *oauth2.Config
	parseFunc  func(ctx context.Context, token string, cli *http.Client) (*dto.ProviderResponse, error)
}

func NewOAuth2Provider(
	clientID,
	clientSecret,
	redirectURL string,
	endpoint oauth2.Endpoint,
	scopes []string,
	successURL string,
	parseFunc func(ctx context.Context, token string, cli *http.Client) (*dto.ProviderResponse, error),
) *OAuth2Provider {
	return &OAuth2Provider{
		successURL: successURL,
		parseFunc:  parseFunc,
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     endpoint,
		},
	}
}

func (o *OAuth2Provider) GetSuccessURL() string {
	return o.successURL
}

func (o *OAuth2Provider) GetConfig() *oauth2.Config {
	return o.config
}

func (o *OAuth2Provider) GetUser(ctx context.Context, code string) (*dto.ProviderResponse, error) {
	const op = "provider.GetUser.google"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	token, err := o.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %v", err)
	}

	cli := o.config.Client(ctx, token)
	res, err := o.parseFunc(ctx, token.AccessToken, cli)
	if err != nil {
		return nil, err
	}

	return &dto.ProviderResponse{
		ProviderID:   res.ProviderID,
		Email:        res.Email,
		Name:         res.Name,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		ExpiresIn:    token.ExpiresIn,
		TokenType:    token.TokenType,
	}, nil
}
