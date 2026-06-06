package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Traefik TraefikConfig `mapstructure:"traefik"`
	Docker  DockerConfig  `mapstructure:"docker"`
}

type TraefikConfig struct {
	URL              string        `mapstructure:"url"`
	PollInterval     time.Duration `mapstructure:"poll_interval"`
	Username         string        `mapstructure:"username"`
	Password         string        `mapstructure:"password"`
	CertificateFiles []string      `mapstructure:"certificate_files"`
}

type DockerConfig struct {
	Socket string `mapstructure:"socket"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName(".porthole")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME")

	v.SetDefault("traefik.url", "http://127.0.0.1:8080")
	v.SetDefault("traefik.poll_interval", 5*time.Second)
	v.SetDefault("docker.socket", "/var/run/docker.sock")

	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
