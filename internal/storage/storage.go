package storage

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/OlegShamkeev/go_final_project/internal/config"
	"github.com/OlegShamkeev/go_final_project/internal/task"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var cfg *config.Config

type Storage struct {
	Db *sqlx.DB
}

func NewStorage(config *config.Config) {
	cfg = config
}

func InitDB(dbPath string) (*sqlx.DB, error) {
	var dbFilePath string
	if len(dbPath) > 0 {
		dbFilePath = dbPath
	} else {
		appPath, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dbFilePath = filepath.Join(appPath, "scheduler.Db")
		log.Printf("Db path that will be used is %s\n", dbFilePath)
	}

	var install bool

	_, err := os.Stat(dbFilePath)

	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Attempt to create new Db file by path: %s\n", dbFilePath)

			install = true

			err = os.MkdirAll(filepath.Dir(dbFilePath), 0766)
			if err != nil {
				return nil, err
			}
			f, err := os.Create(dbFilePath)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			log.Println("New Db file successfully created")
		} else {
			return nil, err
		}
	}
	log.Printf("Connecting to Db by path: %s\n", dbFilePath)
	Db, err := sqlx.Connect("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	if install {
		if err = createTableAndIndex(Db); err != nil {
			return nil, err
		}
	}

	return Db, nil
}

func createTableAndIndex(Db *sqlx.DB) error {
	log.Println("Creating new table scheduler with index")
	schema := `CREATE TABLE scheduler (id INTEGER PRIMARY KEY AUTOINCREMENT, date CHAR(8) NOT NULL DEFAULT "", 
	title VARCHAR(256) NOT NULL DEFAULT "", comment TEXT NOT NULL DEFAULT "", repeat VARCHAR(128) NOT NULL DEFAULT "")`

	_, err := Db.Exec(schema)
	if err != nil {
		return err
	}
	index := `CREATE INDEX scheduler_date ON scheduler (date)`
	_, err = Db.Exec(index)
	if err != nil {
		return err
	}
	return nil
}

func (t Storage) CreateTask(task *task.Task) (int, error) {
	insertRow := `INSERT INTO scheduler (date, title, comment, repeat) 
	VALUES (?, ?, ?, ?)`
	res, err := t.Db.Exec(insertRow, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (t Storage) GetTasks(search string) ([]task.Task, error) {
	var selectRows string
	tasks := []task.Task{}
	var errM error

	switch length := len(search); {
	case length > 0:
		date, err := time.Parse("02.01.2006", search)
		if err != nil {
			selectRows = `SELECT * FROM scheduler WHERE UPPER(title) LIKE ? OR UPPER(comment) LIKE ? ORDER BY date LIMIT ?`
			errM = t.Db.Select(&tasks, selectRows,
				"%"+strings.ToUpper(search)+"%",
				"%"+strings.ToUpper(search)+"%",
				cfg.Limit)
			break
		}
		selectRows = `SELECT * FROM scheduler WHERE date = ? LIMIT ?`
		errM = t.Db.Select(&tasks, selectRows, date.Format("20060102"), cfg.Limit)

	case length == 0:
		selectRows = `SELECT * FROM scheduler ORDER BY date LIMIT ?`
		errM = t.Db.Select(&tasks, selectRows, cfg.Limit)
	}

	if errM != nil {
		return nil, errM
	}
	return tasks, nil
}

func (t Storage) GetTask(id int) (*task.Task, error) {
	task := &task.Task{}
	selectRow := `SELECT * FROM scheduler WHERE id = ?`
	err := t.Db.Get(task, selectRow, id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (t Storage) UpdateTask(task *task.Task) error {
	updateRow := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	_, err := t.Db.Exec(updateRow, task.Date, task.Title, task.Comment, task.Repeat, task.Id)
	if err != nil {
		return err
	}
	return nil
}

func (t Storage) DeleteTask(id int) error {
	deleteRow := `DELETE FROM scheduler where id = ?`
	_, err := t.Db.Exec(deleteRow, id)
	if err != nil {
		return err
	}
	return nil
}
