package main

import (
	"context"
	"github.com/invakid404/magicauth/config"
	"github.com/invakid404/magicauth/http"
	"github.com/invakid404/magicauth/k8s"
	"github.com/invakid404/magicauth/oauth"
	"github.com/invakid404/magicauth/oauth/client"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	cfg := config.Get()

	auth := oauth.New(cfg)
	for id, data := range cfg.OAuthClients {
		result, err := client.ToOAuthClient(auth, id, data)
		if err != nil {
			log.Fatalln("failed to map client", id, "from config:", err)
		}

		auth.UpsertClient(result)
	}

	api := http.New(cfg, auth)

	if cfg.EnableK8S {
		controllerQuit, err := k8s.RunController(auth)
		if err != nil {
			log.Fatalln("failed to run controller:", err)
		}

		defer func() {
			controllerQuit <- struct{}{}
		}()
	}

	go func() {
		log.Fatalln(api.Server.ListenAndServe())
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)
	<-quit

	api.Server.SetKeepAlivesEnabled(false)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_ = api.Server.Shutdown(ctx)
}
