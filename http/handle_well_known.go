package http

import (
	"encoding/base64"
	"encoding/json"
	"github.com/invakid404/magicauth/util/url"
	"math/big"
	"net/http"
)

type wellKnownResponse struct {
	Issuer           string   `json:"issuer"`
	AuthEndpoint     string   `json:"authorization_endpoint"`
	TokenEndpoint    string   `json:"token_endpoint"`
	UserInfoEndpoint string   `json:"userinfo_endpoint"`
	JwksURI          string   `json:"jwks_uri"`
	Algorithms       []string `json:"id_token_signing_alg_values_supported"`
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	N         string `json:"n"`
	E         string `json:"e"`
}

func (h *HTTP) wellKnown(res http.ResponseWriter, req *http.Request) {
	response := wellKnownResponse{
		Issuer:           h.config.BaseURL,
		AuthEndpoint:     url.MustJoinPath(h.config.BaseURL, authPath),
		TokenEndpoint:    url.MustJoinPath(h.config.BaseURL, tokenPath),
		UserInfoEndpoint: url.MustJoinPath(h.config.BaseURL, userinfoPath),
		JwksURI:          url.MustJoinPath(h.config.BaseURL, jwksPath),
		Algorithms:       []string{"RS256"},
	}

	res.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(response)
	_, _ = res.Write(data)
}

func (h *HTTP) jwks(res http.ResponseWriter, req *http.Request) {
	publicKey := &h.oauth.GetPrivateKey().PublicKey

	// Encode the RSA public key as JWK
	jwkKey := jwk{
		KeyID:     h.oauth.GetKeyID(),
		KeyType:   "RSA",
		Use:       "sig",
		Algorithm: "RS256",
		N:         base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
		E:         base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
	}

	response := jwksResponse{
		Keys: []jwk{jwkKey},
	}

	res.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(response)
	_, _ = res.Write(data)
}
