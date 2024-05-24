package http

import (
	"github.com/ory/fosite"
	"net/http"
	"regexp"
)

func (h *HTTP) auth(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	authReq, err := h.oauth.Provider.NewAuthorizeRequest(ctx, req)
	if err != nil {
		h.oauth.Provider.WriteAuthorizeError(ctx, res, authReq, err)

		return
	}

	username := req.Header.Get("Tailscale-User-Login")
	if username == "" {
		h.oauth.Provider.WriteAuthorizeError(ctx, res, authReq, fosite.ErrRequestUnauthorized)
		return
	}

	session := h.oauth.NewSession(replaceSpecialChars(username))

	response, err := h.oauth.Provider.NewAuthorizeResponse(ctx, authReq, session)
	if err != nil {
		h.oauth.Provider.WriteAuthorizeError(ctx, res, authReq, err)
		return
	}

	h.oauth.Provider.WriteAuthorizeResponse(ctx, res, authReq, response)
}

var specialCharRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func replaceSpecialChars(str string) string {
	return specialCharRegex.ReplaceAllString(str, "_")
}
