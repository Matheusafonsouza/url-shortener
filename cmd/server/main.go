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

	"golang-urlshortener/internal/config"
	internalhttp "golang-urlshortener/internal/http"
	"golang-urlshortener/internal/http/api"
	"golang-urlshortener/internal/http/web"
	"golang-urlshortener/internal/shorturl"
	"golang-urlshortener/internal/storage/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	repo := postgres.NewRepository(dbPool)
	service := shorturl.NewService(repo, cfg.DefaultTTL)

	apiHandlers := api.NewHandlers(service, cfg.BaseURL)
	webHandlers := web.NewHandlers(service, cfg.BaseURL)
	redirectHandler := internalhttp.NewRedirectHandler(service)

	templates, err := web.ParseTemplates()
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	router := internalhttp.NewRouter(apiHandlers, webHandlers, redirectHandler.Redirect, templates)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		log.Printf("server running on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	log.Println("server stopped")
}
