package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/OlegShamkeev/go_final_project/internal/config"
	"github.com/OlegShamkeev/go_final_project/internal/database"
	"github.com/OlegShamkeev/go_final_project/internal/transport"
	"github.com/OlegShamkeev/go_final_project/internal/utils"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
)

var cfg config.Config
var store *database.Storage
var secret []byte

func main() {
	var err error
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error during parse enviroment variable(s) %s", err.Error())
		return
	}
	database.Cfg = &cfg
	transport.Cfg = &cfg

	if len(cfg.Password) > 0 {
		secret = utils.GenerateSecret()
		transport.Secret = secret
	}

	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store = &database.Storage{Db: db}
	transport.Store = store

	r := chi.NewRouter()

	r.Handle("/*", http.FileServer(http.Dir(cfg.WebFolder)))

	r.Get("/api/nextdate", transport.GetNextDate)
	r.Post("/api/task", transport.Auth(transport.PostTask))
	r.Get("/api/tasks", transport.Auth(transport.GetTasks))
	r.Get("/api/task", transport.Auth(transport.GetTask))
	r.Put("/api/task", transport.Auth(transport.UpdateTask))
	r.Post("/api/task/done", transport.Auth(transport.CheckDoneTask))
	r.Delete("/api/task", transport.Auth(transport.DeleteTask))
	r.Post("/api/signin", transport.AuthAndGenerateToken)

	log.Printf("Starting web-server on port: %d\n", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.Port), r); err != nil {
		log.Fatalf("Error starting web-server: %s", err.Error())
	}
}
