package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/invakid404/magicauth/config"
	"github.com/invakid404/magicauth/k8s"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
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
	secret      = []byte("my super secret signing password")
	oauthConfig = &fosite.Config{
		AccessTokenLifespan: time.Minute * 60,
		GlobalSecret:        secret,
	}
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	provider      = compose.ComposeAllEnabled(oauthConfig, store, privateKey)
)

var specialCharRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func replaceSpecialChars(str string) string {
	return specialCharRegex.ReplaceAllString(str, "_")
}

func main() {
	cfg := config.Get()

	controllerQuit, err := k8s.RunController()
	if err != nil {
		log.Fatalln("failed to run controller:", err)
	}

	http.HandleFunc("/auth", authEndpoint)
	http.HandleFunc("/token", tokenEndpoint)
	http.HandleFunc("/userinfo", userinfoEndpoint)

	log.Println("listening on", cfg.Port)

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
		Handler: http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			metrics := httpsnoop.CaptureMetrics(http.DefaultServeMux, res, req)

			log.Printf(
				"%s %s (code=%d dt=%s written=%d)",
				req.Method,
				req.URL,
				metrics.Code,
				metrics.Duration,
				metrics.Written,
			)
		}),
	}

	go func() {
		log.Fatalln(server.ListenAndServe())
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)
	<-quit

	server.SetKeepAlivesEnabled(false)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_ = server.Shutdown(ctx)

	controllerQuit <- struct{}{}
}

func authEndpoint(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	authReq, err := provider.NewAuthorizeRequest(ctx, req)
	if err != nil {
		provider.WriteAuthorizeError(ctx, res, authReq, err)

		return
	}

	username := req.Header.Get("Tailscale-User-Login")
	if username == "" {
		provider.WriteAuthorizeError(ctx, res, authReq, fosite.ErrRequestUnauthorized)
		return
	}

	session := newSession(replaceSpecialChars(username))

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
			Issuer:      config.Get().BaseURL,
			Subject:     user,
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
