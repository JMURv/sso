package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Mode        string       `yaml:"mode" env-default:"dev"`
	ServiceName string       `yaml:"serviceName" env-required:"true"`
	Auth        AuthConfig   `yaml:"auth"`
	Server      ServerConfig `yaml:"server"`
	Email       EmailConfig  `yaml:"email"`
	DB          DBConfig     `yaml:"db"`
	Redis       RedisConfig  `yaml:"redis"`
	Jaeger      JaegerConfig `yaml:"jaeger"`
}

type provider struct {
	ClientID     string   `yaml:"clientID"`
	ClientSecret string   `yaml:"clientSecret"`
	RedirectURL  string   `yaml:"redirectURL"`
	Scopes       []string `yaml:"scopes"`
}

type AuthConfig struct {
	Secret string `yaml:"secret" env-required:"true"`
	Oauth  struct {
		Google provider `yaml:"google"`
	} `yaml:"oauth"`

	OIDC struct {
		Google provider `yaml:"google"`
	} `yaml:"oidc"`
}

type ServerConfig struct {
	Port   int    `yaml:"port" env-required:"true"`
	Scheme string `yaml:"scheme" env-default:"http"`
	Domain string `yaml:"domain" env-default:"localhost"`
}

type EmailConfig struct {
	Server string `yaml:"server" env-default:"smtp.gmail.com"`
	Port   int    `yaml:"port" env-default:"587"`
	User   string `yaml:"user" env-default:""`
	Pass   string `yaml:"pass" env-default:""`
	Admin  string `yaml:"admin" env-default:""`
}

type DBConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	Database string `yaml:"database" env-default:"db"`
}

type RedisConfig struct {
	Addr string `yaml:"addr" env-default:"localhost:6379"`
	Pass string `yaml:"pass" env-default:""`
}

type JaegerConfig struct {
	Sampler struct {
		Type  string  `yaml:"type"`
		Param float64 `yaml:"param"`
	} `yaml:"sampler"`
	Reporter struct {
		LogSpans           bool   `yaml:"LogSpans"`
		LocalAgentHostPort string `yaml:"LocalAgentHostPort"`
		CollectorEndpoint  string `yaml:"CollectorEndpoint"`
	} `yaml:"reporter"`
}

func MustLoad(configPath string) Config {
	conf := Config{}

	data, err := os.ReadFile(configPath)
	if err != nil {
		panic("failed to read config: " + err.Error())
	}

	if err = yaml.Unmarshal(data, &conf); err != nil {
		panic("failed to unmarshal config: " + err.Error())
	}

	return conf
}
