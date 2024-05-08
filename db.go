package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type storage struct {
	db *sqlx.DB
}

func openDB() (*sqlx.DB, error) {
	var dbFilePath string
	if len(cfg.DBPath) > 0 {
		dbFilePath = cfg.DBPath
	} else {
		appPath, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dbFilePath = filepath.Join(appPath, "scheduler.db")
		fmt.Printf("I take default DB path %s\n", dbFilePath)
	}

	var install bool

	_, err := os.Stat(dbFilePath)

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("trying create new DB here %s\n", dbFilePath)

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
		} else {
			return nil, err
		}
	}

	db, err := sqlx.Connect("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	if install {
		if err = CreateTableTasks(db); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func CreateTableTasks(db *sqlx.DB) error {
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

func (t storage) createTask(task Task) (int, error) {
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
