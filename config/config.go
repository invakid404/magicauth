package config

import (
	"github.com/caarlos0/env/v11"
	"log"
	"sync"
)

type Config struct {
	Port    int    `env:"MAGICAUTH_PORT" envDefault:"8080"`
	BaseURL string `env:"MAGICAUTH_BASE_URL,required"`
}

var (
	config     Config
	configInit sync.Once
)

func Get() *Config {
	configInit.Do(func() {
		if err := env.Parse(&config); err != nil {
			log.Fatalln(err)
		}
	})

	return &config
}
