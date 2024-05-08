package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

var cfg config
var store *storage

func main() {
	var err error
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error during parse enviroment variable(s) %s", err.Error())
		return
	}

	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store = &storage{db: db}

	r := chi.NewRouter()

	r.Handle("/*", http.FileServer(http.Dir(cfg.WebFolder)))

	r.Get("/api/nextdate", getNextDate)
	r.Post("/api/task", postTask)

	if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.Port), r); err != nil {
		log.Fatalf("Error starting web-server: %s", err.Error())
	}
}
