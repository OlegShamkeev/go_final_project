package services

import (
	"strings"
	"time"

	"github.com/OlegShamkeev/go_final_project/internal/types"
)

type Task struct {
	Id      string `json:"id,omitempty" db:"id"`
	Date    string `json:"date,omitempty" db:"date"`
	Title   string `json:"title" db:"title"`
	Comment string `json:"comment,omitempty" db:"comment"`
	Repeat  string `json:"repeat,omitempty" db:"repeat"`
}

func (task *Task) ValidateAndUpdateTask(update bool) *types.Result {

	if len(strings.TrimSpace(task.Title)) == 0 {
		return &types.Result{Error: "field title couldn't be empty"}
	}

	if len(strings.TrimSpace(task.Date)) == 0 {
		task.Date = time.Now().Format(dateTimeFormat)
	} else {
		dateParsed, err := time.Parse(dateTimeFormat, task.Date)
		if err != nil {
			return &types.Result{Error: err.Error()}
		}
		if len(strings.TrimSpace(task.Repeat)) > 0 {
			task.Date, err = NextDate(time.Now(), task.Date, task.Repeat, update)
			if err != nil {
				return &types.Result{Error: err.Error()}
			}
		} else if dateParsed.Before(time.Now()) {
			task.Date = time.Now().Format(dateTimeFormat)
		}
	}
	return nil
}
