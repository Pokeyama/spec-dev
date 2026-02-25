package main

import (
	"log"
	"net/http"

	"gashasystem/internal/config"
	"gashasystem/internal/persistence"
	"gashasystem/internal/server"
	"gashasystem/internal/session"
)

func main() {
	cfg := config.Load()

	repo, err := persistence.NewRepository(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("failed to close repository: %v", err)
		}
	}()

	sessions := session.NewStore(cfg.MemcachedAddr)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: server.NewMux(repo, sessions, cfg),
	}

	log.Printf("api listening on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
