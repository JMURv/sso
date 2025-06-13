package config

import (
	"github.com/caarlos0/env/v9"
	"go.uber.org/zap"
)

type Config struct {
	Mode        string `env:"MODE" envDefault:"dev"`
	ServiceName string `env:"SERVICE_NAME" envDefault:"sso"`
	Server      ServerConfig
	Auth        authConfig
	Email       smtpConfig
	DB          dbConfig
	Minio       s3Config
	Redis       redisConfig
	Prometheus  prometheusConfig
	Jaeger      jaegerConfig
}

type ServerConfig struct {
	Port     int    `env:"BACKEND_PORT,required"`
	GRPCPort int    `env:"BACKEND_GRPC_PORT" envDefault:"50050"`
	Scheme   string `env:"SERVER_SCHEME" envDefault:"http"`
	Domain   string `env:"SERVER_DOMAIN" envDefault:"localhost"`
}

type authConfig struct {
	Admins []string `env:"ADMIN_USERS" envSeparator:","`

	JWT struct {
		Secret string `env:"JWT_SECRET,required"`
		Issuer string `env:"JWT_ISSUER,required"`
	} `yaml:"jwt"`

	Captcha struct {
		SiteKey string `env:"CAPTCHA_SITE_KEY"`
		Secret  string `env:"CAPTCHA_SECRET"`
	} `yaml:"captcha"`

	WebAuthn struct {
		Origins []string `env:"WEBAUTHN_ORIGINS" envSeparator:","`
	} `yaml:"webauthn"`

	Providers struct {
		Secret     string `env:"PROVIDER_SECRET"`
		SuccessURL string `env:"PROVIDER_SUCCESS_URL"`

		Oauth struct {
			Google struct {
				ClientID     string   `env:"OAUTH2_GOOGLE_CLIENT_ID" envDefault:""`
				ClientSecret string   `env:"OAUTH2_GOOGLE_CLIENT_SECRET" envDefault:""`
				RedirectURL  string   `env:"OAUTH2_GOOGLE_REDIRECT_URL" envDefault:""`
				Scopes       []string `env:"OAUTH2_GOOGLE_SCOPES" envDefault:"" envSeparator:","`
			} `yaml:"google"`
		} `yaml:"oauth"`

		OIDC struct {
			Google struct {
				ClientID     string   `env:"OIDC_GOOGLE_CLIENT_ID" envDefault:""`
				ClientSecret string   `env:"OIDC_GOOGLE_CLIENT_SECRET" envDefault:""`
				RedirectURL  string   `env:"OIDC_GOOGLE_REDIRECT_URL" envDefault:""`
				Scopes       []string `env:"OIDC_GOOGLE_SCOPES" envDefault:"" envSeparator:","`
			} `yaml:"google"`
		} `yaml:"oidc"`
	} `yaml:"providers"`
}

type smtpConfig struct {
	Server string `env:"EMAIL_SERVER" envDefault:"smtp.gmail.com"`
	Port   int    `env:"EMAIL_PORT" envDefault:"587"`
	User   string `env:"EMAIL_USER" envDefault:""`
	Pass   string `env:"EMAIL_PASS" envDefault:""`
	Admin  string `env:"EMAIL_ADMIN" envDefault:""`
}

type dbConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER" envDefault:"postgres"`
	Password string `env:"DB_PASSWORD" envDefault:"postgres"`
	Database string `env:"DB_DATABASE" envDefault:"db"`
}

type s3Config struct {
	Addr      string `env:"MINIO_ADDR" envDefault:"localhost:9000"`
	AccessKey string `env:"MINIO_ACCESS_KEY" envDefault:""`
	SecretKey string `env:"MINIO_SECRET_KEY" envDefault:""`
	Bucket    string `env:"MINIO_BUCKET" envDefault:""`
	UseSSL    bool   `env:"MINIO_SSL" envDefault:"false"`
}

type redisConfig struct {
	Addr string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Pass string `env:"REDIS_PASS" envDefault:""`
}

type prometheusConfig struct {
	Port int `env:"BACKEND_METRICS_PORT" envDefault:"8085"`
}

type jaegerConfig struct {
	Sampler struct {
		Type  string  `env:"JAEGER_SAMPLER_TYPE" envDefault:"const"`
		Param float64 `env:"JAEGER_SAMPLER_PARAM" envDefault:"1"`
	} `yaml:"sampler"`
	Reporter struct {
		LogSpans           bool   `env:"JAEGER_REPORTER_LOGSPANS" envDefault:"true"`
		LocalAgentHostPort string `env:"JAEGER_REPORTER_LOCALAGENT" envDefault:"localhost:6831"`
		CollectorEndpoint  string `env:"JAEGER_REPORTER_COLLECTOR" envDefault:"http://localhost:14268/api/traces"`
	} `yaml:"reporter"`
}

func MustLoad() Config {
	conf := Config{}
	if err := env.Parse(&conf); err != nil {
		panic("failed to parse environment variables: " + err.Error())
	}

	zap.L().Info("Load configuration from environment")
	return conf
}
