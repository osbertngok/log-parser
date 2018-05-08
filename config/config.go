package config

import (
	"github.com/caarlos0/env"
)

type Config struct {
	Timezone string `env:"TIMEZONE" envDefault:"Asia/Hong_Kong"`
}

func New() *Config {
	conf := &Config{}
	conf.Load()
	return conf
}

func (c *Config) Load() {
	err := env.Parse(c)
	if err != nil {
		panic(err)
	}
}
