package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"log"
	"os"
	"strings"
	"sync"
)

type Config struct {
	Port         int                    `koanf:"port"`
	BaseURL      string                 `koanf:"base_url" validate:"required"`
	EnableK8S    bool                   `koanf:"enable_k8s"`
	GlobalSecret string                 `koanf:"global_secret" validate:"required"`
	OAuthClients map[string]OAuthClient `koanf:"clients"`
}

type OAuthClient map[string]any

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

		// Load environment variables (config file takes precedence)
		err := k.Load(env.ProviderWithValue("MAGICAUTH_", ".", func(key, value string) (string, any) {
			key = strings.ReplaceAll(
				strings.ToLower(
					strings.TrimPrefix(key, "MAGICAUTH_"),
				),
				"__", ".",
			)

			var newValue any = value
			if strings.Contains(value, ",") {
				newValue = strings.Split(value, ",")
			}

			return key, newValue
		}), nil)

		if err != nil {
			log.Fatalln(fmt.Errorf("failed to load env variables: %w", err))
		}

		// Check for config files
		configFiles := []string{"config.json", "config.yaml", "config.yml", "config.toml"}
		var foundFiles []string
		for _, cf := range configFiles {
			if _, err := os.Stat(cf); err == nil {
				foundFiles = append(foundFiles, cf)
			}
		}

		if len(foundFiles) > 1 {
			log.Fatalln("multiple configuration files found, provide only one")
		}

		// Load config file if found
		if len(foundFiles) == 1 {
			target := foundFiles[0]

			var parser koanf.Parser
			switch {
			case strings.HasSuffix(target, ".json"):
				parser = json.Parser()
			case strings.HasSuffix(target, ".yaml") || strings.HasSuffix(target, ".yml"):
				parser = yaml.Parser()
			case strings.HasSuffix(target, ".toml"):
				parser = toml.Parser()
			}

			if err := k.Load(file.Provider(target), parser); err != nil {
				log.Fatalf("error loading config file: %v", err)
			}
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
