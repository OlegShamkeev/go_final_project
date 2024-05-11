package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type storage struct {
	db *sqlx.DB
}

func initDB() (*sqlx.DB, error) {
	var dbFilePath string
	if len(cfg.DBPath) > 0 {
		dbFilePath = cfg.DBPath
	} else {
		appPath, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dbFilePath = filepath.Join(appPath, "scheduler.db")
		log.Printf("DB path that will be used is %s\n", dbFilePath)
	}

	var install bool

	_, err := os.Stat(dbFilePath)

	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Attempt to create new DB file by path: %s\n", dbFilePath)

			install = true
			f, err := os.Create(dbFilePath)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			err = f.Chmod(0666)
			if err != nil {
				return nil, err
			}
			log.Println("New DB file successfully created")
		} else {
			return nil, err
		}
	}
	log.Printf("Connecting to DB by path: %s\n", dbFilePath)
	db, err := sqlx.Connect("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	if install {
		if err = createTableAndIndex(db); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func createTableAndIndex(db *sqlx.DB) error {
	log.Println("Creating new table scheduler with index")
	schema := `CREATE TABLE scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT, date CHAR(8) NOT NULL DEFAULT "", 
	title VARCHAR(256) NOT NULL DEFAULT "", comment TEXT NOT NULL DEFAULT "", repeat VARCHAR(128) NOT NULL DEFAULT "")`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}
	index := `CREATE INDEX scheduler_date ON scheduler (date)`
	_, err = db.Exec(index)
	if err != nil {
		return err
	}
	return nil
}

func (t storage) createTask(task *Task) (int, error) {
	log.Printf("Insert new record in DB:\n date: %s, title: %s, comment: %s, repeat: %s\n", task.Date, task.Title, task.Comment, task.Repeat)
	insertRow := `INSERT INTO scheduler (date, title, comment, repeat) 
	VALUES (?, ?, ?, ?)`
	res, err := t.db.Exec(insertRow, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (t storage) getTasks(search string) ([]Task, error) {
	var selectRows string
	tasks := []Task{}
	var errM error

	if len(search) > 0 {
		date, err := time.Parse("02.01.2006", search)
		if err != nil {
			selectRows = `SELECT * FROM scheduler WHERE UPPER(title) LIKE ? OR UPPER(comment) LIKE ? ORDER BY date LIMIT ?`
			errM = t.db.Select(&tasks, selectRows,
				"%"+strings.ToUpper(search)+"%",
				"%"+strings.ToUpper(search)+"%",
				cfg.Limit)
		} else {
			selectRows = `SELECT * FROM scheduler WHERE date = ? LIMIT ?`
			errM = t.db.Select(&tasks, selectRows, date.Format("20060102"), cfg.Limit)
		}
	} else {
		selectRows = `SELECT * FROM scheduler ORDER BY date LIMIT ?`
		errM = t.db.Select(&tasks, selectRows, cfg.Limit)
	}
	if errM != nil {
		return nil, errM
	}
	return tasks, nil
}

func (t storage) getTask(id int) (*Task, error) {
	task := &Task{}

	selectRow := `SELECT * FROM scheduler WHERE id = ?`
	err := t.db.Get(task, selectRow, id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (t storage) updateTask(task *Task) error {
	log.Printf("Update task with id: %s", task.Id)

	updateRow := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	_, err := t.db.Exec(updateRow, task.Date, task.Title, task.Comment, task.Repeat, task.Id)
	if err != nil {
		return err
	}
	return nil
}
