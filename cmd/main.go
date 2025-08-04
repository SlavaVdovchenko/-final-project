package main

import (
	"log/slog"
	"os"
	"todo_list/internal/app"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)

	if err := app.Run(); err != nil {
		slog.Error("main run dowm", slog.Any("error", err))
		os.Exit(1)
	}
}
