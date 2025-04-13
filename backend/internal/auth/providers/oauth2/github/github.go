package providers

import (
	"context"
	"errors"
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

var ErrNoEmailFound = errors.New("no email found")

type GitHubProvider struct {
	*providers.OAuth2Provider
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
				zap.L().Debug("Failed to create request", zap.Error(err))
				return nil, err
			}
			req.Header.Set("Authorization", "token "+token)

			resp, err := cli.Do(req)
			if err != nil {
				zap.L().Error(
					"Failed to get user",
					zap.String("op", op),
					zap.Error(err),
				)
				return nil, err
			}
			defer func(Body io.ReadCloser) {
				if err := Body.Close(); err != nil {
					zap.L().Debug("failed to close body", zap.Error(err))
				}
			}(resp.Body)

			var ghRes dto.GitHubResponse
			if err = json.NewDecoder(resp.Body).Decode(&ghRes); err != nil {
				zap.L().Error(
					"Failed to decode response",
					zap.String("op", op),
					zap.Error(err),
				)
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
				Picture:    ghRes.AvatarURL,
			}, nil
		},
	)
	return &GitHubProvider{provider}
}

func getGitHubEmail(ctx context.Context, cli *http.Client, token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		zap.L().Debug("Failed to create request", zap.Error(err))
		return "", err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := cli.Do(req)
	if err != nil {
		zap.L().Debug("Failed to get user emails", zap.Error(err))
		return "", err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			zap.L().Debug("failed to close body", zap.Error(err))
		}
	}(resp.Body)

	var res map[string]any
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return "", ErrNoEmailFound
}
