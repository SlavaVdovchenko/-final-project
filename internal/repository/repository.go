package repository

import (
	"database/sql"
	"fmt"
	"log/slog"
	"todo_list/internal/models"

	"github.com/pkg/errors"
)

type ListRepo struct {
	conn *sql.DB
}

func New(dbConn *sql.DB) *ListRepo {
	return &ListRepo{
		conn: dbConn,
	}
}

func (l *ListRepo) AddTask(request models.ListDB) (int, error) {
	res, err := l.conn.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		request.Date.Format("20060102"), request.Title, request.Comment, request.Repeat)
	if err != nil {
		return 0, errors.Wrap(err, "adding to db")
	}

	ID, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "last insert")
	}

	return int(ID), nil
}

func (l *ListRepo) GetLastTasks(limit int) ([]*models.List, error) {
	result := make([]*models.List, 0, limit)
	selectString := fmt.Sprintf("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT %d", limit)
	rows, err := l.conn.Query(selectString)
	if err != nil {
		return nil, errors.Wrap(err, "select last lists")
	}

	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("can not close rows", slog.Any("error", err))
		}
	}()

	for rows.Next() {
		var id int
		var date, title, comment, repeat string

		if err := rows.Scan(&id, &date, &title, &comment, &repeat); err != nil {
			return nil, errors.Wrap(err, "rows scan")
		}

		result = append(result, &models.List{
			ID:      id,
			Date:    date,
			Title:   title,
			Comment: comment,
			Repeat:  repeat,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows err")
	}

	return result, nil
}

func (l *ListRepo) GetTaskByID(id string) (models.List, error) {
	var result models.List

	row := l.conn.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id =?", id)

	if err := row.Scan(&result.ID, &result.Date, &result.Title, &result.Comment, &result.Repeat); err != nil {
		return result, errors.Wrap(err, "get task by id")
	}

	return result, nil
}

func (l *ListRepo) UpdateTask(task models.ListDB) error {
	_, err := l.conn.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?", task.Date.Format("20060102"),
		task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return errors.Wrap(err, "update task")
	}
	return nil
}

func (l *ListRepo) UpdateTaskDate(task models.ListDB) error {
	_, err := l.conn.Exec("UPDATE scheduler SET date = ? WHERE id = ?", task.Date.Format("20060102"), task.ID)
	if err != nil {
		return errors.Wrap(err, "update task date")
	}
	return nil
}

func (l *ListRepo) DeleteTask(id string) error {
	_, err := l.conn.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return errors.Wrap(err, "delete task")
	}
	return nil
}
