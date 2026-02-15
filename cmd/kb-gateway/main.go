package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HardMakabaka/KB-Gateway/internal/api"
	"github.com/HardMakabaka/KB-Gateway/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           api.NewServer(cfg),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("kb-gateway listening on %s", cfg.HTTP.Addr)
		errCh <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Printf("shutdown signal: %s", sig)
	case err := <-errCh:
		log.Printf("server error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
