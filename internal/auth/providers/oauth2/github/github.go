package providers

import (
	"context"
	"fmt"
	providers "github.com/JMURv/sso/internal/auth/providers/oauth2"
	conf "github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/goccy/go-json"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/oauth2/github"
	"io"
	"net/http"
	"strconv"
)

type GitHubProvider struct {
	*providers.OAuth2Provider
}

type gitHubResponse struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func NewGitHubOAuth2(config conf.Config) *GitHubProvider {
	provider := providers.NewOAuth2Provider(
		config.Auth.Oauth.GitHub.ClientID,
		config.Auth.Oauth.GitHub.ClientSecret,
		config.Auth.Oauth.GitHub.RedirectURL,
		github.Endpoint,
		config.Auth.Oauth.GitHub.Scopes,
		config.Auth.Oauth.SuccessURL,
		func(ctx context.Context, token string, cli *http.Client) (*dto.ProviderResponse, error) {
			const op = "provider.GetUser.github"
			span, ctx := opentracing.StartSpanFromContext(ctx, op)
			defer span.Finish()

			req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", "token "+token)

			resp, err := cli.Do(req)
			if err != nil {
				return nil, err
			}

			defer func(Body io.ReadCloser) {
				if err := Body.Close(); err != nil {
					zap.L().Debug("failed to close body", zap.Error(err))
				}
			}(resp.Body)

			var ghRes gitHubResponse
			if err = json.NewDecoder(resp.Body).Decode(&ghRes); err != nil {
				return nil, err
			}

			if ghRes.Email == "" {
				email, err := getGitHubEmail(ctx, cli, token)
				if err != nil {
					return nil, err
				}
				ghRes.Email = email
			}

			return &dto.ProviderResponse{
				ProviderID: strconv.FormatInt(ghRes.ID, 10),
				Email:      ghRes.Email,
				Name:       ghRes.Name,
			}, nil
		},
	)
	return &GitHubProvider{provider}
}

func getGitHubEmail(ctx context.Context, cli *http.Client, token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+token)

	resp, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			zap.L().Debug("failed to close body", zap.Error(err))
		}
	}(resp.Body)

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}
	return "", fmt.Errorf("no verified email found")
}
