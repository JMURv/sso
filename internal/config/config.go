package config

import (
	"errors"
	"github.com/caarlos0/env/v9"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Mode        string       `yaml:"mode" env:"MODE" envDefault:"dev"`
	ServiceName string       `yaml:"serviceName" env:"SERVICE_NAME" envDefault:"sso"`
	Auth        AuthConfig   `yaml:"auth"`
	Server      ServerConfig `yaml:"server"`
	Email       EmailConfig  `yaml:"email"`
	DB          DBConfig     `yaml:"db"`
	Redis       RedisConfig  `yaml:"redis"`
	Jaeger      JaegerConfig `yaml:"jaeger"`
}

type AuthConfig struct {
	Secret             string `yaml:"secret" env:"SECRET,required"`
	ProviderSignSecret string `yaml:"providerSignSecret" env:"PROVIDER_SIGN_SECRET"`
	Oauth              struct {
		SuccessURL string `yaml:"successURL" env:"OAUTH2_SUCCESS_URL"`
		Google     struct {
			ClientID     string   `yaml:"clientID" env:"OAUTH2_GOOGLE_CLIENT_ID" envDefault:""`
			ClientSecret string   `yaml:"clientSecret" env:"OAUTH2_GOOGLE_CLIENT_SECRET" envDefault:""`
			RedirectURL  string   `yaml:"redirectURL" env:"OAUTH2_GOOGLE_REDIRECT_URL" envDefault:""`
			Scopes       []string `yaml:"scopes" env:"OAUTH2_GOOGLE_SCOPES" envDefault:"" envSeparator:","`
		} `yaml:"google"`
		GitHub struct {
			ClientID     string   `yaml:"clientID" env:"OAUTH2_GITHUB_CLIENT_ID" envDefault:""`
			ClientSecret string   `yaml:"clientSecret" env:"OAUTH2_GITHUB_CLIENT_SECRET" envDefault:""`
			RedirectURL  string   `yaml:"redirectURL" env:"OAUTH2_GITHUB_REDIRECT_URL" envDefault:""`
			Scopes       []string `yaml:"scopes" env:"OAUTH2_GITHUB_SCOPES" envDefault:"" envSeparator:","`
		} `yaml:"github"`
	} `yaml:"oauth"`

	OIDC struct {
		SuccessURL string `yaml:"successURL" env:"OIDC_SUCCESS_URL"`
		Google     struct {
			ClientID     string   `yaml:"clientID" env:"OIDC_GOOGLE_CLIENT_ID" envDefault:""`
			ClientSecret string   `yaml:"clientSecret" env:"OIDC_GOOGLE_CLIENT_SECRET" envDefault:""`
			RedirectURL  string   `yaml:"redirectURL" env:"OIDC_GOOGLE_REDIRECT_URL" envDefault:""`
			Scopes       []string `yaml:"scopes" env:"OIDC_GOOGLE_SCOPES" envDefault:"" envSeparator:","`
		} `yaml:"google"`
	} `yaml:"oidc"`
}

type ServerConfig struct {
	Port     int    `yaml:"port" env:"SERVER_PORT,required"`
	GRPCPort int    `yaml:"grpc_port" env:"SERVER_GRPC_PORT" envDefault:"50065"`
	Scheme   string `yaml:"scheme" env:"SERVER_SCHEME" envDefault:"http"`
	Domain   string `yaml:"domain" env:"SERVER_DOMAIN" envDefault:"localhost"`
}

type EmailConfig struct {
	Server string `yaml:"server" env:"EMAIL_SERVER" envDefault:"smtp.gmail.com"`
	Port   int    `yaml:"port" env:"EMAIL_PORT" envDefault:"587"`
	User   string `yaml:"user" env:"EMAIL_USER" envDefault:""`
	Pass   string `yaml:"pass" env:"EMAIL_PASS" envDefault:""`
	Admin  string `yaml:"admin" env:"EMAIL_ADMIN" envDefault:""`
}

type DBConfig struct {
	Host     string `yaml:"host" env:"DB_HOST" envDefault:"localhost"`
	Port     int    `yaml:"port" env:"DB_PORT" envDefault:"5432"`
	User     string `yaml:"user" env:"DB_USER" envDefault:"postgres"`
	Password string `yaml:"password" env:"DB_PASSWORD" envDefault:"postgres"`
	Database string `yaml:"database" env:"DB_DATABASE" envDefault:"db"`
}

type RedisConfig struct {
	Addr string `yaml:"addr" env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Pass string `yaml:"pass" env:"REDIS_PASS" envDefault:""`
}

type JaegerConfig struct {
	Sampler struct {
		Type  string  `yaml:"type" env:"JAEGER_SAMPLER_TYPE" envDefault:"const"`
		Param float64 `yaml:"param" env:"JAEGER_SAMPLER_PARAM" envDefault:"1"`
	} `yaml:"sampler"`
	Reporter struct {
		LogSpans           bool   `yaml:"LogSpans" env:"JAEGER_REPORTER_LOGSPANS" envDefault:"true"`
		LocalAgentHostPort string `yaml:"LocalAgentHostPort" env:"JAEGER_REPORTER_LOCALAGENT" envDefault:"localhost:6831"`
		CollectorEndpoint  string `yaml:"CollectorEndpoint" env:"JAEGER_REPORTER_COLLECTOR" envDefault:"http://localhost:14268/api/traces"`
	} `yaml:"reporter"`
}

func MustLoad(configPath string) Config {
	conf := Config{}

	_, err := os.Stat(configPath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		if err = env.Parse(conf); err != nil {
			panic("failed to parse environment variables: " + err.Error())
		}

		zap.L().Info(
			"Load configuration from environment",
		)
		return conf
	} else if err != nil {
		panic("failed to stat file: " + err.Error())
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		panic("failed to read config: " + err.Error())
	}

	if err = yaml.Unmarshal(data, &conf); err != nil {
		panic("failed to unmarshal config: " + err.Error())
	}

	zap.L().Info(
		"Load configuration from yaml",
		zap.String("path", configPath),
	)
	return conf
}
