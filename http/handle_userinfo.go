package http

import (
	"encoding/json"
	"fmt"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"net/http"
)

func (h *HTTP) userinfo(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	session := h.oauth.NewSession("")
	tokenType, accessReq, err := h.oauth.Provider.IntrospectToken(ctx, fosite.AccessTokenFromRequest(req), fosite.AccessToken, session)
	if err != nil {
		h.oauth.Provider.WriteIntrospectionError(ctx, res, err)
		return
	}

	if tokenType != fosite.AccessToken {
		h.oauth.Provider.WriteIntrospectionError(ctx, res, fosite.ErrInvalidTokenFormat)
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
