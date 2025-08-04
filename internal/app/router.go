package app

import (
	"net/http"
	"todo_list/internal/handler"
	"todo_list/internal/usecase"
)

func getRouter(uc *usecase.TaskUC) http.Handler {
	mux := http.NewServeMux()

	h := handler.New(uc)

	mux.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.AddTaskHandler(w, r)
		case http.MethodGet:
			h.GetTaskhandler(w, r)
		case http.MethodPut:
			h.PutTaskHandler(w, r)
		case http.MethodDelete:
			h.DeleteTaskHandler(w, r)
		}
	})

	mux.HandleFunc("/api/nextdate", h.GetNextDatehandler)

	mux.HandleFunc("/api/task/done", h.PutDoneTaskHandler)

	mux.HandleFunc("/api/tasks", h.GetLastTasksHandler)

	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)
	return mux
}
