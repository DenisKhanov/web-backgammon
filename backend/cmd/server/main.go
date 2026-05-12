package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/denis/web-backgammon/internal/api"
	"github.com/denis/web-backgammon/internal/config"
	"github.com/denis/web-backgammon/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect to DB", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, cfg.MigrationsDir); err != nil {
		slog.Error("run migrations", "err", err)
		os.Exit(1)
	}

	srv := api.NewServer(
		db.NewRoomRepo(pool),
		db.NewPlayerRepo(pool),
		db.NewGameRepo(pool),
		cfg.AllowedOrigins,
	)

	addr := ":" + cfg.Port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
