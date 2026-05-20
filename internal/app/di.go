package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	chatapi "mygochat/internal/api/chat"
	"mygochat/internal/api/health"
	"mygochat/internal/censor"
	"mygochat/internal/chat"
	"mygochat/internal/config"
	"mygochat/internal/migrations"
	messagerepo "mygochat/internal/repository/message"
)

type diContainer struct {
	cfg    *config.Config
	logger *slog.Logger

	db          *pgxpool.Pool
	hub         *chat.Hub
	censor      *censor.Censor
	chatSvc     *chat.Service
	chatHandler *chatapi.Handler
}

func newDIContainer(cfg *config.Config, logger *slog.Logger) *diContainer {
	return &diContainer{cfg: cfg, logger: logger}
}

func (c *diContainer) DB() *pgxpool.Pool {
	if c.db != nil {
		return c.db
	}

	poolCfg, err := pgxpool.ParseConfig(c.cfg.DB.DSN)
	if err != nil {
		c.logger.Error("failed to parse db dsn", slog.String("error", err.Error()))
		os.Exit(1)
	}
	poolCfg.MaxConns = c.cfg.DB.MaxConns
	poolCfg.MinConns = c.cfg.DB.MinConns
	poolCfg.MaxConnLifetime = c.cfg.DB.MaxConnLifetime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		c.logger.Error("failed to create db pool", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := pool.Ping(context.Background()); err != nil {
		c.logger.Error("failed to ping db", slog.String("error", err.Error()))
		os.Exit(1)
	}

	c.logger.Info("database connected")
	c.db = pool
	return c.db
}

func (c *diContainer) Migrate() {
	db := stdlib.OpenDB(*c.DB().Config().ConnConfig)
	defer func() {
		if err := db.Close(); err != nil {
			c.logger.Error("failed to close migration db", slog.String("error", err.Error()))
		}
	}()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		c.logger.Error("failed to set goose dialect", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := goose.UpContext(context.Background(), db, "."); err != nil {
		c.logger.Error("failed to apply migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}
	c.logger.Info("migrations applied")
}

func (c *diContainer) Hub() *chat.Hub {
	if c.hub == nil {
		c.hub = chat.NewHub(c.logger)
	}
	return c.hub
}

func (c *diContainer) Censor() *censor.Censor {
	if c.censor == nil {
		c.censor = censor.New(c.cfg.Chat.Profanities)
	}
	return c.censor
}

func (c *diContainer) ChatService() *chat.Service {
	if c.chatSvc == nil {
		repo := messagerepo.New(c.DB())
		c.chatSvc = chat.NewService(repo, c.Censor(), c.Hub(), c.cfg.Chat.HistoryLimit)
	}
	return c.chatSvc
}

func (c *diContainer) ChatHandler() *chatapi.Handler {
	if c.chatHandler == nil {
		c.chatHandler = chatapi.NewHandler(
			c.ChatService(),
			c.Hub(),
			c.logger,
			c.cfg.Chat,
			c.cfg.App.TrustedOrigins,
		)
	}
	return c.chatHandler
}

func (c *diContainer) HealthHandler() *health.Handler {
	return health.NewHandler(config.Version)
}
