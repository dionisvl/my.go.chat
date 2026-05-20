package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"mygochat/internal/api/middleware"
	"mygochat/internal/api/root"
	"mygochat/internal/config"
)

type App struct {
	cfg     *config.Config
	logger  *slog.Logger
	di      *diContainer
	httpSrv *http.Server
}

func New() *App {
	cfg := config.Load()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	a := &App{
		cfg:    cfg,
		logger: logger,
		di:     newDIContainer(cfg, logger),
	}

	a.initDeps()

	a.httpSrv = &http.Server{
		Addr:              cfg.App.Port,
		Handler:           a.buildRouter(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return a
}

func (a *App) initDeps() {
	a.di.DB()
	a.di.Migrate()
}

func (a *App) buildRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(a.logger))
	r.Use(chimiddleware.Recoverer)

	r.Get("/", root.ServeHTTP)
	r.Get("/health", a.di.HealthHandler().ServeHTTP)
	r.Get("/ws", a.di.ChatHandler().ServeHTTP)

	return r
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		a.logger.Info("starting server",
			slog.String("addr", a.cfg.App.Port),
			slog.String("env", a.cfg.App.Env),
		)
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.App.ShutdownTimeout)
	defer cancel()

	if err := a.httpSrv.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		return err
	}

	a.di.DB().Close()
	a.logger.Info("server stopped cleanly")
	return nil
}
