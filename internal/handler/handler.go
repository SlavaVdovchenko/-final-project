package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"todo_list/internal/models"
	"todo_list/internal/usecase"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Handle struct {
	uc *usecase.ListUC
}

func New(usecase *usecase.ListUC) *Handle {
	return &Handle{
		uc: usecase,
	}
}

func parseRepeat(repeat string) (string, int, error) {
	if repeat == "" {
		return "", 0, nil
	}

	parts := strings.Split(repeat, " ")
	if len(parts) == 0 {
		return "", 0, errors.New("empty repeat rule")
	}

	rule := parts[0]
	switch rule {
	case "y":
		if len(parts) != 1 {
			return "", 0, errors.New("invalid format y rule")
		}
		return "y", 0, nil
	case "d":
		if len(parts) != 2 {
			return "", 0, errors.New("invalid format d rule")
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n <= 0 || n > 400 {
			return "", 0, errors.New("invalid format d rule")
		}
		return "d", n, nil
	default:
		return "", 0, errors.New("unsupported repeat rule")
	}

}

func writeError(w http.ResponseWriter, status int, msg string, err error) {
	slog.Error(msg, slog.Any("error", err))
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": msg,
	})
}
func (h *Handle) AddTaskHandler(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "read request body", err)
		return
	}
	defer req.Body.Close()

	var taskRequest models.List

	if err := json.Unmarshal(body, &taskRequest); err != nil {
		writeError(w, http.StatusBadRequest, "unmarshal json", err)
		return
	}

	if taskRequest.Date == "" || taskRequest.Date == "today" {
		taskRequest.Date = time.Now().Format("20060102")
	}

	if err := validate.Struct(taskRequest); err != nil {
		writeError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	date, err := time.Parse("20060102", taskRequest.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date format (YYYYMMDD required)", err)
		return
	}

	if taskRequest.Repeat == "" {
		now := time.Now().Truncate(24 * time.Hour)
		if date.Before(now) {
			writeError(w, http.StatusBadRequest, "date cannot be in the past for non-repeating tasks", nil)
			return
		}
	}

	rpt, interval, err := parseRepeat(taskRequest.Repeat)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid repeat rule", err)
		return
	}

	task := models.ListDB{
		Date:       date,
		Title:      taskRequest.Title,
		Comment:    taskRequest.Comment,
		Repeat:     taskRequest.Repeat,
		RepeatRule: rpt,
		Interval:   interval,
	}

	var response models.ListPostResponse
	response.ID, err = h.uc.Add(task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add task", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", slog.Any("error", err))
		return
	}
}

func (h *Handle) GetLastTasksHandler(w http.ResponseWriter, req *http.Request) {
	result, err := h.uc.GetLast(10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get last tasks", err)
		return
	}

	tasks := make([]map[string]string, len(result))
	for i, item := range result {
		tasks[i] = map[string]string{
			"id":      strconv.Itoa(item.ID),
			"date":    item.Date,
			"title":   item.Title,
			"comment": item.Comment,
			"repeat":  item.Repeat,
		}
	}

	response := map[string]interface{}{
		"tasks": tasks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", slog.Any("error", err))
		return
	}
}

func (h *Handle) GetTaskhandler(w http.ResponseWriter, req *http.Request) {
	idStr := req.URL.Query().Get("id")
	if idStr == "" {
		slog.Error("ID required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "ID required",
		})
		return
	}

	response, err := h.uc.GetByID(idStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "task not found", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", slog.Any("error", err))
		return
	}
}

func (h *Handle) PutTaskHandler(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "read request body", err)
		return
	}
	defer req.Body.Close()

	var taskRequest models.List

	if err := json.Unmarshal(body, &taskRequest); err != nil {
		writeError(w, http.StatusBadRequest, "unmarshal json", err)
		return
	}

	if taskRequest.Date == "" || taskRequest.Date == "today" {
		taskRequest.Date = time.Now().Format("20060102")
	}

	if err := validate.Struct(taskRequest); err != nil {
		writeError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	date, err := time.Parse("20060102", taskRequest.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date format (YYYYMMDD required)", err)
		return
	}

	if taskRequest.Repeat == "" {
		now := time.Now().Truncate(24 * time.Hour)
		if date.Before(now) {
			writeError(w, http.StatusBadRequest, "date cannot be in the past for non-repeating tasks", nil)
			return
		}
	}

	rpt, interval, err := parseRepeat(taskRequest.Repeat)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid repeat rule", err)
		return
	}

	task := models.ListDB{
		ID:         taskRequest.ID,
		Date:       date,
		Title:      taskRequest.Title,
		Comment:    taskRequest.Comment,
		Repeat:     taskRequest.Repeat,
		RepeatRule: rpt,
		Interval:   interval,
	}

	err = h.uc.Update(task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "task not found", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(models.ListPutResponse{ID: task.ID}); err != nil {
		slog.Error("failed to encode response", slog.Any("error", err))
		return
	}
}

func (h *Handle) PutDoneTaskHandler(w http.ResponseWriter, req *http.Request) {
	idStr := req.URL.Query().Get("id")
	if idStr == "" {
		slog.Error("ID required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "ID required",
		})
		return
	}

	response, err := h.uc.GetByID(idStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "task not found", err)
		return
	}

	if response.Repeat == "" {
		if err := h.uc.Delete(idStr); err != nil {
			writeError(w, http.StatusInternalServerError, "delete faild", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
		return
	}

	date, err := time.Parse("20060102", response.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date format (YYYYMMDD required)", err)
		return
	}

	rpt, interval, err := parseRepeat(response.Repeat)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid repeat rule", err)
		return
	}

	task := models.ListDB{
		ID:         response.ID,
		Date:       date,
		Title:      response.Title,
		Comment:    response.Comment,
		Repeat:     response.Repeat,
		RepeatRule: rpt,
		Interval:   interval,
	}

	if err := h.uc.Done(task); err != nil {
		writeError(w, http.StatusInternalServerError, "delete faild", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})

}

func (h *Handle) DeleteTaskHandler(w http.ResponseWriter, req *http.Request) {
	idStr := req.URL.Query().Get("id")
	if idStr == "" {
		slog.Error("ID required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "ID required",
		})
		return
	}

	_, err := h.uc.GetByID(idStr)
	if err != nil {
		writeError(w, http.StatusNotFound, "task not found", err)
		return
	}

	if err := h.uc.Delete(idStr); err != nil {
		writeError(w, http.StatusInternalServerError, "delete faild", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}
