package usecase

import (
	"strconv"
	"strings"
	"time"
	"todo_list/internal/models"
	"todo_list/internal/repository"
	"todo_list/tests"

	"github.com/pkg/errors"
)

type TaskUC struct {
	TaskRepo *repository.TaskRepo
}

func New(list *repository.TaskRepo) *TaskUC {
	return &TaskUC{
		TaskRepo: list,
	}
}

func afterNow(date, now time.Time) bool {
	if date.After(now) {
		return true
	}
	return false
}

func (l *TaskUC) nextDate(now time.Time, dstart string, repeat string) (string, error) {
	date, err := time.Parse(tests.DateParsingFormat, dstart)
	if err != nil {
		return "", errors.Wrap(err, "next date")
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", nil
	}

	switch parts[0] {
	case "d":
		days, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", errors.Wrap(err, "parsing d rule")
		}
		for {
			date = date.AddDate(0, 0, days)
			if afterNow(date, now) {
				break
			}
		}
	case "y":
		for {
			date = date.AddDate(1, 0, 0)
			if afterNow(date, now) {
				break
			}
		}
	default:
		return "", errors.New("incorrect repeat rule")

	}
	return date.Format(tests.DateParsingFormat), nil
}

func (l *TaskUC) Add(request models.Task) (int, error) {
	return l.TaskRepo.AddTask(request)
}

func (l *TaskUC) GetLast(n int) ([]*models.Task, error) {
	return l.TaskRepo.GetLastTasks(n)
}

func (l *TaskUC) GetByID(id string) (models.Task, error) {
	return l.TaskRepo.GetTaskByID(id)
}

func (l *TaskUC) Update(task models.Task) error {
	return l.TaskRepo.UpdateTask(task)
}

func (l *TaskUC) Done(task models.Task) error {
	date, err := l.nextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		return errors.Wrap(err, "next date")
	}

	task.Date = date

	return l.TaskRepo.UpdateTaskDate(task)
}

func (l *TaskUC) Delete(id string) error {
	return l.TaskRepo.DeleteTask(id)
}

func (l *TaskUC) NextDate(ask models.Task, now time.Time) (string, error) {
	return l.nextDate(now, ask.Date, ask.Repeat)
}
