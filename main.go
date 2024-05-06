package main

import (
	"fmt"
	"net/http"

	"github.com/caarlos0/env"
)

var cfg config

func main() {

	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("Error during parse enviroment variable(s) %s", err.Error())
	}

	http.Handle("/", http.FileServer(http.Dir(cfg.WebFolder)))

	if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.Port), nil); err != nil {
		fmt.Printf("Error starting web-server: %s", err.Error())
	}
}
