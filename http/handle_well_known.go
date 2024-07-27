package http

import (
	"encoding/json"
	"github.com/invakid404/magicauth/util/url"
	"net/http"
)

type wellKnownResponse struct {
	Issuer           string   `json:"issuer"`
	AuthEndpoint     string   `json:"authorization_endpoint"`
	TokenEndpoint    string   `json:"token_endpoint"`
	UserInfoEndpoint string   `json:"userinfo_endpoint"`
	Algorithms       []string `json:"id_token_signing_alg_values_supported"`
}

func (h *HTTP) wellKnown(res http.ResponseWriter, req *http.Request) {
	response := wellKnownResponse{
		Issuer:           h.config.BaseURL,
		AuthEndpoint:     url.MustJoinPath(h.config.BaseURL, authPath),
		TokenEndpoint:    url.MustJoinPath(h.config.BaseURL, tokenPath),
		UserInfoEndpoint: url.MustJoinPath(h.config.BaseURL, userinfoPath),
		Algorithms:       []string{"RS256"},
	}

	data, _ := json.Marshal(response)
	_, _ = res.Write(data)
}
