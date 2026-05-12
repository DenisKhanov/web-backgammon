package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/denis/web-backgammon/internal/api"
	"github.com/denis/web-backgammon/internal/config"
	"github.com/denis/web-backgammon/internal/db"
	"github.com/denis/web-backgammon/internal/ws"
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

	rooms := db.NewRoomRepo(pool)
	players := db.NewPlayerRepo(pool)
	games := db.NewGameRepo(pool)

	hub := ws.NewHub(ws.DBRepos{
		Rooms:   rooms,
		Players: players,
		Games:   games,
	}, cfg.AllowedOrigins)

	srv := api.NewServer(rooms, players, games, cfg.AllowedOrigins, hub)

	addr := ":" + cfg.Port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
