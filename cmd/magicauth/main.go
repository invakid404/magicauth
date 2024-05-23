package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"
	"log"
	"net/http"
	"time"
)

var (
	store = &storage.MemoryStore{
		IDSessions: make(map[string]fosite.Requester),
		Clients: map[string]fosite.Client{
			"my-client": &fosite.DefaultClient{
				ID:            "my-client",
				Secret:        []byte(`$2a$10$IxMdI6d.LIRZPpSfEwNoeu4rY3FhDREsxFJXikcgdRRAStxUlsuEO`), // = "foobar"
				RedirectURIs:  []string{"http://localhost:3000/auth/oidc.callback"},
				ResponseTypes: []string{"id_token", "code", "token", "id_token token", "code id_token", "code token", "code id_token token"},
				GrantTypes:    []string{"implicit", "refresh_token", "authorization_code", "password", "client_credentials"},
				Scopes:        []string{"openid"},
			},
		},
		AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
		AccessTokens:           map[string]fosite.Requester{},
		RefreshTokens:          map[string]storage.StoreRefreshToken{},
		PKCES:                  map[string]fosite.Requester{},
		AccessTokenRequestIDs:  map[string]string{},
		RefreshTokenRequestIDs: map[string]string{},
		IssuerPublicKeys:       map[string]storage.IssuerPublicKeys{},
		PARSessions:            map[string]fosite.AuthorizeRequester{},
	}
	secret = []byte("my super secret signing password")
	config = &fosite.Config{
		AccessTokenLifespan: time.Minute * 60,
		GlobalSecret:        secret,
	}
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	provider      = compose.ComposeAllEnabled(config, store, privateKey)
)

func main() {
	http.HandleFunc("/auth", authEndpoint)
	http.HandleFunc("/token", tokenEndpoint)
	http.HandleFunc("/userinfo", userinfoEndpoint)

	log.Fatalln(http.ListenAndServe(":8080", nil))
}

func authEndpoint(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	authReq, err := provider.NewAuthorizeRequest(ctx, req)
	if err != nil {
		provider.WriteAuthorizeError(ctx, res, authReq, err)

		return
	}

	session := newSession("gosho")

	response, err := provider.NewAuthorizeResponse(ctx, authReq, session)
	if err != nil {
		provider.WriteAuthorizeError(ctx, res, authReq, err)
		return
	}

	provider.WriteAuthorizeResponse(ctx, res, authReq, response)
}

func tokenEndpoint(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	session := newSession("")
	accessReq, err := provider.NewAccessRequest(ctx, req, session)

	if err != nil {
		provider.WriteAccessError(ctx, res, accessReq, err)
		return
	}

	if accessReq.GetGrantTypes().ExactOne("client_credentials") {
		for _, scope := range accessReq.GetRequestedScopes() {
			accessReq.GrantScope(scope)
		}
	}

	response, err := provider.NewAccessResponse(ctx, accessReq)
	if err != nil {
		provider.WriteAccessError(ctx, res, accessReq, err)
		return
	}

	provider.WriteAccessResponse(ctx, res, accessReq, response)
}

func userinfoEndpoint(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	session := newSession("")
	tokenType, accessReq, err := provider.IntrospectToken(ctx, fosite.AccessTokenFromRequest(req), fosite.AccessToken, session)
	if err != nil {
		provider.WriteIntrospectionError(ctx, res, err)
		return
	}

	if tokenType != fosite.AccessToken {
		provider.WriteIntrospectionError(ctx, res, fosite.ErrInvalidTokenFormat)
		return
	}

	session = accessReq.GetSession().(*openid.DefaultSession)

	response := map[string]any{
		"sub":   session.Claims.Subject,
		"email": fmt.Sprintf("%s@tailscale.com", session.Claims.Subject),
	}

	data, _ := json.Marshal(response)
	_, _ = res.Write(data)
}

func newSession(user string) *openid.DefaultSession {
	return &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Issuer:      "http://localhost:8080",
			Subject:     user,
			Audience:    []string{"http://localhost:3000"},
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
