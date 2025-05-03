package config

import (
	"errors"
	"github.com/caarlos0/env/v9"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Mode        string           `yaml:"mode" env:"MODE" envDefault:"dev"`
	ServiceName string           `yaml:"service_name" env:"SERVICE_NAME" envDefault:"sso"`
	Server      ServerConfig     `yaml:"server"`
	Auth        authConfig       `yaml:"auth"`
	Email       smtpConfig       `yaml:"email"`
	DB          dbConfig         `yaml:"db"`
	Minio       s3Config         `yaml:"minio"`
	Redis       redisConfig      `yaml:"redis"`
	Prometheus  prometheusConfig `yaml:"prometheus"`
	Jaeger      jaegerConfig     `yaml:"jaeger"`
}

type ServerConfig struct {
	Port     int    `yaml:"port" env:"BACKEND_PORT,required"`
	GRPCPort int    `yaml:"grpc_port" env:"BACKEND_GRPC_PORT" envDefault:"50050"`
	Scheme   string `yaml:"scheme" env:"SERVER_SCHEME" envDefault:"http"`
	Domain   string `yaml:"domain" env:"SERVER_DOMAIN" envDefault:"localhost"`
}

type authConfig struct {
	Admins []string `yaml:"admins" env:"ADMIN_USERS" envSeparator:","`

	JWT struct {
		Secret string `yaml:"secret" env:"JWT_SECRET,required"`
		Issuer string `yaml:"issuer" env:"JWT_ISSUER,required"`
	} `yaml:"jwt"`

	Captcha struct {
		SiteKey string `yaml:"site_key" env:"CAPTCHA_SITE_KEY"`
		Secret  string `yaml:"secret" env:"CAPTCHA_SECRET"`
	} `yaml:"captcha"`

	WebAuthn struct {
		Origins []string `yaml:"origins" env:"WEBAUTHN_ORIGINS" envSeparator:","`
	} `yaml:"webauthn"`

	Providers struct {
		Secret     string `yaml:"secret" env:"PROVIDER_SECRET"`
		SuccessURL string `yaml:"success_url" env:"PROVIDER_SUCCESS_URL"`

		Oauth struct {
			Google struct {
				ClientID     string   `yaml:"client_id" env:"OAUTH2_GOOGLE_CLIENT_ID" envDefault:""`
				ClientSecret string   `yaml:"client_secret" env:"OAUTH2_GOOGLE_CLIENT_SECRET" envDefault:""`
				RedirectURL  string   `yaml:"redirect_url" env:"OAUTH2_GOOGLE_REDIRECT_URL" envDefault:""`
				Scopes       []string `yaml:"scopes" env:"OAUTH2_GOOGLE_SCOPES" envDefault:"" envSeparator:","`
			} `yaml:"google"`
		} `yaml:"oauth"`

		OIDC struct {
			Google struct {
				ClientID     string   `yaml:"client_id" env:"OIDC_GOOGLE_CLIENT_ID" envDefault:""`
				ClientSecret string   `yaml:"client_secret" env:"OIDC_GOOGLE_CLIENT_SECRET" envDefault:""`
				RedirectURL  string   `yaml:"redirect_url" env:"OIDC_GOOGLE_REDIRECT_URL" envDefault:""`
				Scopes       []string `yaml:"scopes" env:"OIDC_GOOGLE_SCOPES" envDefault:"" envSeparator:","`
			} `yaml:"google"`
		} `yaml:"oidc"`
	} `yaml:"providers"`
}

type smtpConfig struct {
	Server string `yaml:"server" env:"EMAIL_SERVER" envDefault:"smtp.gmail.com"`
	Port   int    `yaml:"port" env:"EMAIL_PORT" envDefault:"587"`
	User   string `yaml:"user" env:"EMAIL_USER" envDefault:""`
	Pass   string `yaml:"pass" env:"EMAIL_PASS" envDefault:""`
	Admin  string `yaml:"admin" env:"EMAIL_ADMIN" envDefault:""`
}

type dbConfig struct {
	Host     string `yaml:"host" env:"DB_HOST" envDefault:"localhost"`
	Port     int    `yaml:"port" env:"DB_PORT" envDefault:"5432"`
	User     string `yaml:"user" env:"DB_USER" envDefault:"postgres"`
	Password string `yaml:"password" env:"DB_PASSWORD" envDefault:"postgres"`
	Database string `yaml:"database" env:"DB_DATABASE" envDefault:"db"`
}

type s3Config struct {
	Addr      string `yaml:"addr" env:"MINIO_ADDR" envDefault:"localhost:9000"`
	AccessKey string `yaml:"access_key" env:"MINIO_ACCESS_KEY" envDefault:""`
	SecretKey string `yaml:"secret_key" env:"MINIO_SECRET_KEY" envDefault:""`
	Bucket    string `yaml:"bucket" env:"MINIO_BUCKET" envDefault:""`
	UseSSL    bool   `yaml:"use_ssl" env:"MINIO_SSL" envDefault:"false"`
}

type redisConfig struct {
	Addr string `yaml:"addr" env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Pass string `yaml:"pass" env:"REDIS_PASS" envDefault:""`
}

type prometheusConfig struct {
	Port int `yaml:"port" env:"BACKEND_METRICS_PORT" envDefault:"8085"`
}

type jaegerConfig struct {
	Sampler struct {
		Type  string  `yaml:"type" env:"JAEGER_SAMPLER_TYPE" envDefault:"const"`
		Param float64 `yaml:"param" env:"JAEGER_SAMPLER_PARAM" envDefault:"1"`
	} `yaml:"sampler"`
	Reporter struct {
		LogSpans           bool   `yaml:"log_spans" env:"JAEGER_REPORTER_LOGSPANS" envDefault:"true"`
		LocalAgentHostPort string `yaml:"local_agent_host_port" env:"JAEGER_REPORTER_LOCALAGENT" envDefault:"localhost:6831"`
		CollectorEndpoint  string `yaml:"collector_endpoint" env:"JAEGER_REPORTER_COLLECTOR" envDefault:"http://localhost:14268/api/traces"`
	} `yaml:"reporter"`
}

func MustLoad(configPath string) Config {
	conf := Config{}

	_, err := os.Stat(configPath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		if err = env.Parse(&conf); err != nil {
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
