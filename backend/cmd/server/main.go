// Сервер для VPS / docker compose.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	_ "time/tzdata"

	"expensebot/config"
	"expensebot/internal/app/sqlite/conn"
	"expensebot/server"
	// часовые пояса без системной tzdata (distroless-образ)
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		return err
	}

	db, err := conn.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := conn.Migrate(ctx, db); err != nil {
		return err
	}

	app, err := server.New(cfg, server.Deps{DB: db, HTTPClient: &http.Client{Timeout: 90 * time.Second}, Log: log})
	if err != nil {
		return err
	}

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: app.HTTP, ReadHeaderTimeout: 10 * time.Second}

	errs := make(chan error, 2)
	go func() {
		log.Info("http server listening", "addr", cfg.HTTPAddr, "telegram", cfg.Telegram.Mode, "llm", cfg.LLM.Provider, "db", cfg.DBPath)

		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	pollerDone := make(chan struct{})
	go func() {
		defer close(pollerDone)

		if cfg.Telegram.Mode == config.TelegramPolling {
			if err := server.NewPoller(app, log).Run(ctx); err != nil {
				errs <- fmt.Errorf("telegram polling: %w", err)
			}
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutting down")
	case err := <-errs:
		stop()
		<-pollerDone

		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)

	<-pollerDone // дожидаемся ответов на уже полученные сообщения

	return err
}
