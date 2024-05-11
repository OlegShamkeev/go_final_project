package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
)

var cfg config
var store *storage

func main() {
	var err error
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error during parse enviroment variable(s) %s", err.Error())
		return
	}

	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store = &storage{db: db}

	r := chi.NewRouter()

	r.Handle("/*", http.FileServer(http.Dir(cfg.WebFolder)))

	r.Get("/api/nextdate", getNextDate)
	r.Post("/api/task", postTask)
	r.Get("/api/tasks", getTasks)
	r.Get("/api/task", getTask)
	r.Put("/api/task", updateTask)

	log.Printf("Starting web-server on port: %d\n", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.Port), r); err != nil {
		log.Fatalf("Error starting web-server: %s", err.Error())
	}
}
