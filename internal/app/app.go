package app

import (
	"log/slog"
	"net/http"
	"strconv"
	"todo_list/internal/repository"
	"todo_list/internal/storage"
	"todo_list/internal/usecase"
	"todo_list/tests"

	"github.com/pkg/errors"
)

func Run() error {
	conn, err := storage.GetConnect()
	if err != nil {
		return errors.Wrap(err, "get connect")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("close db", slog.Any("error", err))
		}
	}()

	repo := repository.New(conn)
	uc := usecase.New(repo)
	handler := getRouter(uc)

	if err := http.ListenAndServe(":"+strconv.Itoa(tests.Port), handler); err != nil {
		return errors.Wrap(err, "app.Run down")
	}
	return nil
}
