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
var secret []byte

func main() {
	var err error
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error during parse enviroment variable(s) %s", err.Error())
		return
	}

	if len(cfg.Password) > 0 {
		secret = generateSecret()
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
	r.Post("/api/task", auth(postTask))
	r.Get("/api/tasks", auth(getTasks))
	r.Get("/api/task", auth(getTask))
	r.Put("/api/task", auth(updateTask))
	r.Post("/api/task/done", auth(checkDoneTask))
	r.Delete("/api/task", auth(deleteTask))
	r.Post("/api/signin", authAndGenerateToken)

	log.Printf("Starting web-server on port: %d\n", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.Port), r); err != nil {
		log.Fatalf("Error starting web-server: %s", err.Error())
	}
}
