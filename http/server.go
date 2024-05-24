package http

import (
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/invakid404/magicauth/config"
	"github.com/invakid404/magicauth/oauth"
	"log"
	"net/http"
)

type HTTP struct {
	Server *http.Server
	oauth  *oauth.OAuth
	config *config.Config
}

func New(cfg *config.Config, oauth *oauth.OAuth) *HTTP {
	mux := http.NewServeMux()

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
		Handler: http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			metrics := httpsnoop.CaptureMetrics(mux, res, req)

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

	h := &HTTP{
		Server: server,
		oauth:  oauth,
		config: cfg,
	}

	mux.HandleFunc("/auth", h.auth)
	mux.HandleFunc("/token", h.token)
	mux.HandleFunc("/userinfo", h.userinfo)

	return h
}
