package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/caarlos0/env"
)

var cfg config

func main() {

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error during parse enviroment variable(s) %s", err.Error())
		return
	}

	var dbFilePath string
	if len(cfg.DBPath) > 0 {
		dbFilePath = cfg.DBPath
	} else {
		appPath, err := os.Executable()
		if err != nil {
			log.Fatal(err)
			return
		}
		dbFilePath = filepath.Join(filepath.Dir(appPath), "scheduler.db")
	}

	_, err := NewTaskStore(dbFilePath)
	if err != nil {
		log.Fatal(err)
		return
	}

	http.Handle("/", http.FileServer(http.Dir(cfg.WebFolder)))

	if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.Port), nil); err != nil {
		log.Fatalf("Error starting web-server: %s", err.Error())
	}
}
