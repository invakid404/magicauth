package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"log"
	"strings"
	"sync"
)

type Config struct {
	Port      int    `koanf:"port"`
	BaseURL   string `koanf:"base_url" validate:"required"`
	EnableK8S bool   `koanf:"enable_k8s"`
}

var (
	config = Config{
		Port:      8080,
		EnableK8S: false,
	}
	configInit sync.Once
)

func Get() *Config {
	configInit.Do(func() {
		k := koanf.New(".")

		err := k.Load(env.Provider("MAGICAUTH_", ".", func(s string) string {
			return strings.ReplaceAll(
				strings.ToLower(
					strings.TrimPrefix(s, "MAGICAUTH_"),
				),
				"__", ".",
			)
		}), nil)

		if err != nil {
			log.Fatalln(fmt.Errorf("failed to load env variables: %w", err))
		}

		err = k.Unmarshal("", &config)
		if err != nil {
			log.Fatalln(fmt.Errorf("failed to unmarshal config: %w", err))
		}

		validate := validator.New(validator.WithRequiredStructEnabled())
		err = validate.Struct(config)
		if err != nil {
			log.Fatalln(fmt.Errorf("failed to validate config: %w", err))
		}
	})

	return &config
}
