package main

import (
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type storage struct {
	db *sqlx.DB
}

func openDB(path string) (*storage, error) {
	var install bool

	_, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			install = true
			if _, err = os.Create(path); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	db, err := sqlx.Connect("sqlite3", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	store := &storage{db: db}

	if install {
		if err = store.createTableTasks(); err != nil {
			return nil, err
		}
	}

	return store, nil
}

func (t storage) createTableTasks() error {
	_, err := t.db.Exec(`CREATE TABLE scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT, date CHAR(8) NOT NULL DEFAULT "", 
	title VARCHAR(256) NOT NULL DEFAULT "", comment TEXT NOT NULL DEFAULT "", repeat VARCHAR(128) NOT NULL DEFAULT "")`)
	if err != nil {
		return err
	}
	_, err = t.db.Exec(`CREATE INDEX scheduler_date ON scheduler (date)`)
	if err != nil {
		return err
	}
	return nil
}
