package oauth

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/invakid404/magicauth/config"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"
	"log"
	"time"
)

type OAuth struct {
	Provider fosite.OAuth2Provider
	memory   *storage.MemoryStore
	config   *config.Config
}

func New(cfg *config.Config) *OAuth {
	memory := &storage.MemoryStore{
		IDSessions:             make(map[string]fosite.Requester),
		Clients:                map[string]fosite.Client{},
		AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
		AccessTokens:           map[string]fosite.Requester{},
		RefreshTokens:          map[string]storage.StoreRefreshToken{},
		PKCES:                  map[string]fosite.Requester{},
		AccessTokenRequestIDs:  map[string]string{},
		RefreshTokenRequestIDs: map[string]string{},
		IssuerPublicKeys:       map[string]storage.IssuerPublicKeys{},
		PARSessions:            map[string]fosite.AuthorizeRequester{},
	}

	secret := []byte(cfg.GlobalSecret)

	oauthConfig := &fosite.Config{
		AccessTokenLifespan: time.Minute * 60,
		GlobalSecret:        secret,
	}

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	provider := compose.ComposeAllEnabled(oauthConfig, memory, privateKey)

	return &OAuth{
		Provider: provider,
		memory:   memory,
		config:   cfg,
	}
}

func (o *OAuth) UpsertClient(client fosite.Client) {
	o.memory.Clients[client.GetID()] = client
	log.Println("upserted oauth client:", client.GetID())
}

func (o *OAuth) DeleteClient(id string) {
	delete(o.memory.Clients, id)
	log.Println("deleted oauth client:", id)
}

func (o *OAuth) NewSession(username string) *openid.DefaultSession {
	return &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Issuer:      o.config.BaseURL,
			Subject:     username,
			ExpiresAt:   time.Now().Add(time.Hour * 6),
			IssuedAt:    time.Now(),
			RequestedAt: time.Now(),
			AuthTime:    time.Now(),
		},
		Headers: &jwt.Headers{
			Extra: make(map[string]interface{}),
		},
	}
}
