package application

import (
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	PollIntervalSeconds int    `env:"SMS_POLL_INTERVAL_SECONDS" env-default:"60"`
	ModemBaseURL        string `env:"SMS_MODEM_BASE_URL" env-default:"http://192.168.8.1"`
	ModemUsername       string `env:"SMS_MODEM_USERNAME" env-default:"admin"`
	ModemPassword       string `env:"SMS_MODEM_PASSWORD" env-default:"admin"`
	APIBaseURL          string `env:"SMS_API_BASE_URL" env-default:""`
	APIKey              string `env:"SMS_API_KEY" env-default:""`
	LogLevel            string `env:"SMS_LOG_LEVEL" env-default:"info"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return cfg, err
	}
	cfg.APIBaseURL = strings.TrimSuffix(cfg.APIBaseURL, "/")
	return cfg, nil
}
