package http

import "net/http"

func (h *HTTP) token(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	session := h.oauth.NewSession("")
	accessReq, err := h.oauth.Provider.NewAccessRequest(ctx, req, session)

	if err != nil {
		h.oauth.Provider.WriteAccessError(ctx, res, accessReq, err)
		return
	}

	if accessReq.GetGrantTypes().ExactOne("client_credentials") {
		for _, scope := range accessReq.GetRequestedScopes() {
			accessReq.GrantScope(scope)
		}
	}

	response, err := h.oauth.Provider.NewAccessResponse(ctx, accessReq)
	if err != nil {
		h.oauth.Provider.WriteAccessError(ctx, res, accessReq, err)
		return
	}

	h.oauth.Provider.WriteAccessResponse(ctx, res, accessReq, response)
}
