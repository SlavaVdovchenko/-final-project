package main

import (
	"log/slog"
	"os"
	"todo_list/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		slog.Error("main run dowm", slog.Any("error", err))
		os.Exit(1)
	}
}
