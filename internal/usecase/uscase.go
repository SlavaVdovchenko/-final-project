package usecase

import (
	"time"
	"todo_list/internal/models"
	"todo_list/internal/repository"

	"github.com/pkg/errors"
)

type ListUC struct {
	ListRepo *repository.ListRepo
}

func New(list *repository.ListRepo) *ListUC {
	return &ListUC{
		ListRepo: list,
	}
}

func afterNow(date, now time.Time) bool {
	if date.After(now) {
		return true
	}
	return false
}

func (l *ListUC) nextDate(in models.ListDB, now time.Time) (time.Time, error) {
	date := in.Date

	switch in.RepeatRule {
	case "d":
		for {
			date = date.AddDate(0, 0, in.Interval)
			if afterNow(date, now) {
				break
			}
		}
		return date, nil
	case "y":
		date = date.AddDate(1, 0, 0)
		return date, nil
	default:
		return time.Time{}, errors.New("incorrect rule")
	}
}

func (l *ListUC) Add(request models.ListDB) (int, error) {
	return l.ListRepo.AddTask(request)
}

func (l *ListUC) GetLast(n int) ([]*models.List, error) {
	return l.ListRepo.GetLastTasks(n)
}

func (l *ListUC) GetByID(id string) (models.List, error) {
	return l.ListRepo.GetTaskByID(id)
}

func (l *ListUC) Update(task models.ListDB) error {
	return l.ListRepo.UpdateTask(task)
}

func (l *ListUC) Done(task models.ListDB) error {
	date, err := l.nextDate(task, time.Now())
	if err != nil {
		return errors.Wrap(err, "next date")
	}

	task.Date = date

	return l.ListRepo.UpdateTaskDate(task)
}

func (l *ListUC) Delete(id string) error {
	return l.ListRepo.DeleteTask(id)
}
