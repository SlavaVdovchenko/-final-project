package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"todo_list/internal/models"
	"todo_list/internal/usecase"
	"todo_list/tests"
)

type Handle struct {
	uc *usecase.TaskUC
}

func New(usecase *usecase.TaskUC) *Handle {
	return &Handle{
		uc: usecase,
	}
}

func (h *Handle) checkDate(task *models.Task) error {
	now := time.Now()
	today := now.Format(tests.DateParsingFormat)

	enteredToday := task.Date == "today" || task.Date == "" || task.Date == today

	if enteredToday {
		task.Date = today
		return nil
	}

	t, err := time.Parse(tests.DateParsingFormat, task.Date)
	if err != nil {
		return errors.New("invalid date format (YYYYMMDD required)")
	}

	if task.Repeat == "" {
		if t.Before(now) {
			task.Date = today
		}
		return nil
	}

	if !enteredToday && t.Before(now) {
		next, err := h.uc.NextDate(*task, now)
		if err != nil {
			return err
		}
		task.Date = next
	}

	return nil
}

func writeError(w http.ResponseWriter, status int, msg string, err error) {
	slog.Error(msg, slog.Any("error", err))
	writeJSONResponse(w, map[string]string{"error": msg}, status)
}

func writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *Handle) AddTaskHandler(w http.ResponseWriter, req *http.Request) {
	var task models.Task
	body, err := io.ReadAll(req.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed read request body", err)
		return
	}
	defer req.Body.Close()

	if err := json.Unmarshal(body, &task); err != nil {
		writeError(w, http.StatusBadRequest, "ivalid JSON", err)
		return
	}

	if task.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required", nil)
		return
	}

	if err := h.checkDate(&task); err != nil {
		writeError(w, http.StatusBadRequest, "check date", err)
		return
	}

	id, err := h.uc.Add(task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "add task failed", err)
		return
	}

	writeJSONResponse(w, map[string]int{"id": id}, http.StatusCreated)
}

func (h *Handle) GetLastTasksHandler(w http.ResponseWriter, req *http.Request) {
	result, err := h.uc.GetLast(tests.TaskLimit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get last tasks failed", err)
		return
	}

	tasks := make([]models.TaskAPI, 0, len(result))
	for _, t := range result {
		tasks = append(tasks, models.TaskAPI{
			ID:      fmt.Sprintf("%d", t.ID),
			Date:    t.Date,
			Title:   t.Title,
			Comment: t.Comment,
			Repeat:  t.Repeat,
		})
	}

	writeJSONResponse(w, map[string]any{"tasks": tasks}, http.StatusOK)
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

	task, err := h.uc.GetByID(idStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "task not found", err)
		return
	}

	response := models.TaskAPI{
		ID:      fmt.Sprintf("%d", task.ID),
		Date:    task.Date,
		Title:   task.Title,
		Comment: task.Comment,
		Repeat:  task.Repeat,
	}

	writeJSONResponse(w, response, http.StatusOK)
}

func (h *Handle) PutTaskHandler(w http.ResponseWriter, req *http.Request) {
	var input models.TaskAPI
	body, err := io.ReadAll(req.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed read request body", err)
		return
	}
	defer req.Body.Close()

	if err := json.Unmarshal(body, &input); err != nil {
		writeError(w, http.StatusBadRequest, "ivalid JSON", err)
		return
	}

	task, err := h.uc.GetByID(input.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "task not found", err)
		return
	}

	intID, err := strconv.Atoi(input.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "converting id", err)
		return
	}

	task = models.Task{
		ID:      intID,
		Date:    input.Date,
		Title:   input.Title,
		Comment: input.Comment,
		Repeat:  input.Repeat,
	}

	if task.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required", nil)
		return
	}

	if err := h.checkDate(&task); err != nil {
		writeError(w, http.StatusBadRequest, "check date", err)
		return
	}

	err = h.uc.Update(task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update failed", err)
		return
	}

	writeJSONResponse(w, map[string]int{"id": task.ID}, http.StatusOK)
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
			writeError(w, http.StatusInternalServerError, "delete failed", err)
			return
		}

		writeJSONResponse(w, map[string]any{}, http.StatusOK)
		return
	}

	if err := h.uc.Done(response); err != nil {
		writeError(w, http.StatusInternalServerError, "delete faild", err)
		return
	}

	writeJSONResponse(w, map[string]any{}, http.StatusOK)
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
		writeError(w, http.StatusInternalServerError, "delete failеd", err)
		return
	}

	writeJSONResponse(w, map[string]any{}, http.StatusOK)
}

func (h *Handle) GetNextDatehandler(w http.ResponseWriter, r *http.Request) {
	dateStr := r.FormValue("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "date is required", nil)
		return
	}

	repeatStr := r.FormValue("repeat")
	if repeatStr == "" {
		writeError(w, http.StatusBadRequest, "repeat is required", nil)
		return
	}

	nowStr := r.FormValue("now")

	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		parsed, err := time.Parse(tests.DateParsingFormat, nowStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format (YYYYMMDD required)", err)
			return
		}
		now = parsed
	}

	task := models.Task{
		Date:   dateStr,
		Repeat: repeatStr,
	}

	nextDate, err := h.uc.NextDate(task, now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "next date", err)
		return
	}

	writeJSONResponse(w, map[string]string{"next_date": nextDate}, http.StatusOK)
}
