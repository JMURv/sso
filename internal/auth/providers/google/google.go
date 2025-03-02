package google

import (
	"context"
	"fmt"
	conf "github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/goccy/go-json"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
)

type GoogleProvider struct {
	config *oauth2.Config
}

func New(config conf.Config) *GoogleProvider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     config.Auth.Oauth.Google.ClientID,
			ClientSecret: config.Auth.Oauth.Google.ClientSecret,
			RedirectURL:  config.Auth.Oauth.Google.CallbackURL,
			Scopes:       config.Auth.Oauth.Google.Scopes,
			Endpoint:     google.Endpoint,
		},
	}
}

type googleResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Picture    string `json:"picture"`
	VerifEmail bool   `json:"verified_email"`
}

func (g *GoogleProvider) GetConfig() *oauth2.Config {
	return g.config
}

func (g *GoogleProvider) GetUser(ctx context.Context, code string) (*dto.ProviderResponse, error) {
	const op = "provider.GetUser.google"
	span, ctx := opentracing.StartSpanFromContext(ctx, op)
	defer span.Finish()

	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %v", err)
	}

	cli := g.config.Client(ctx, token)
	resp, err := cli.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
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
		ProviderID:   gRes.ID,
		Email:        gRes.Email,
		Name:         gRes.Name,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		ExpiresIn:    token.ExpiresIn,
		TokenType:    token.TokenType,
	}, nil
}
