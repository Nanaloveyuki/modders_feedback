package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"modders-feedback/server/internal/api"
	"modders-feedback/server/internal/auth"
	"modders-feedback/server/internal/config"
	"modders-feedback/server/internal/store"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--healthcheck" {
		healthcheck()
		return
	}
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.DataDir, 0o750); err != nil {
		return err
	}
	db, err := store.Open(cfg.DataDir)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.SeedAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db.StartBackups(ctx)

	server := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: api.New(db, auth.New(cfg.JWTSecret, cfg.CookieSecure), cfg.WebDir).Routes(),
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()
	log.Printf("feedback server listening on %s", cfg.ListenAddr)
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
func healthcheck() {
	response, err := http.Get("http://127.0.0.1" + listenHost() + "/api/health")
	if err != nil {
		os.Exit(1)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}

func listenHost() string {
	if value := os.Getenv("LISTEN_ADDR"); value != "" {
		return value
	}
	return ":8080"
}
