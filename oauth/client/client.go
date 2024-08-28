package client

import (
	"context"
	"fmt"
	"github.com/invakid404/magicauth/oauth"
	"github.com/mitchellh/mapstructure"
	"github.com/ory/fosite"
	"github.com/stoewer/go-strcase"
	"strings"
)

func makeFositeHasher(oauth *oauth.OAuth) func(secret string) ([]byte, error) {
	hasher := oauth.Provider.(*fosite.Fosite).Config.GetSecretsHasher(context.Background())

	return func(secret string) ([]byte, error) {
		hashed, err := hasher.Hash(context.Background(), []byte(secret))
		if err != nil {
			return nil, fmt.Errorf("failed to hash secret: %w", err)
		}

		return hashed, nil
	}
}

func ToOAuthClient(oauth *oauth.OAuth, id string, data map[string]any) (*fosite.DefaultClient, error) {
	hash := makeFositeHasher(oauth)
	var err error

	// Convert keys to snake case
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}

	for _, key := range keys {
		newKey := strcase.SnakeCase(key)
		if key == newKey {
			continue
		}

		data[newKey] = data[key]
		delete(data, key)
	}

	// Hash secrets
	for key, value := range data {
		if !strings.Contains(key, "secret") {
			continue
		}

		if values, ok := value.([]any); ok {
			hashedSecrets := make([][]byte, len(values))
			for idx, value := range values {
				hashedSecrets[idx], err = hash(value.(string))
				if err != nil {
					return nil, fmt.Errorf("failed to hash secret %s: %w", key, err)
				}
			}
			data[key] = hashedSecrets

			continue
		}

		data[key], err = hash(value.(string))
		if err != nil {
			return nil, fmt.Errorf("failed to hash secret %s: %w", key, err)
		}
	}

	// Environment variable convenience stuff:
	// Allow passing string arrays as a single value
	for _, key := range []string{"audience", "redirect_uris", "response_types", "grant_types", "scopes"} {
		if value, ok := data[key].(string); ok {
			data[key] = []string{value}
		}
	}

	// Allow passing boolean values as strings
	for _, key := range []string{"public"} {
		if value, ok := data[key].(string); ok {
			data[key] = value == "true"
		}
	}

	var client fosite.DefaultClient
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &client,
		TagName:  "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create mapstructure decoder: %w", err)
	}

	if err = decoder.Decode(data); err != nil {
		return nil, fmt.Errorf("failed to map oauth client: %w", err)
	}

	client.ID = id

	return &client, nil
}
