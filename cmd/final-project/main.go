package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/OlegShamkeev/go_final_project/internal/api"
	"github.com/OlegShamkeev/go_final_project/internal/config"
	"github.com/OlegShamkeev/go_final_project/internal/storage"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
)

var cfg config.Config

func main() {
	var err error
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error during parse enviroment variable(s) %s", err.Error())
		return
	}

	storage.NewStorage(&cfg)

	db, err := storage.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	api.NewApi(&cfg, &storage.Storage{Db: db})

	r := chi.NewRouter()

	r.Handle("/*", http.FileServer(http.Dir(cfg.WebFolder)))

	r.Get("/api/nextdate", api.GetNextDate)
	r.Post("/api/task", api.Auth(api.PostTask))
	r.Get("/api/tasks", api.Auth(api.GetTasks))
	r.Get("/api/task", api.Auth(api.GetTask))
	r.Put("/api/task", api.Auth(api.UpdateTask))
	r.Post("/api/task/done", api.Auth(api.CheckDoneTask))
	r.Delete("/api/task", api.Auth(api.DeleteTask))
	r.Post("/api/signin", api.AuthAndGenerateToken)

	log.Printf("Starting web-server on port: %d\n", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.Port), r); err != nil {
		log.Fatalf("Error starting web-server: %s", err.Error())
	}
}
