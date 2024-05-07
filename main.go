package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

	_, err := openDB(dbFilePath)
	if err != nil {
		log.Fatal(err)
		return
	}

	http.Handle("/", http.FileServer(http.Dir(cfg.WebFolder)))
	http.HandleFunc("/api/nextdate", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application-json")

		dNow, err := time.Parse("20060102", r.URL.Query().Get("now"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		date := r.URL.Query().Get("date")
		repeat := r.URL.Query().Get("repeat")

		result, err := NextDate(dNow, date, repeat)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		b := []byte(result)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(b)
		if err != nil {
			return
		}
	})

	if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.Port), nil); err != nil {
		log.Fatalf("Error starting web-server: %s", err.Error())
	}
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	d, err := time.Parse("20060102", date)
	if err != nil {
		return "", err
	}

	arrayParams := strings.Split(repeat, " ")

	switch arrayParams[0] {
	case "y":
		for {
			d = d.AddDate(1, 0, 0)
			if d.After(now) {
				break
			}
		}
		return d.Format("20060102"), nil
	case "d":
		if len(arrayParams) <= 1 {
			return "", fmt.Errorf("incorrect format of the repeat parameter")
		}
		daysToAdd, err := strconv.Atoi(arrayParams[1])
		if err != nil {
			return "", err
		}
		if daysToAdd > 400 {
			return "", fmt.Errorf("days to add more than 400")
		}
		for {
			d = d.AddDate(0, 0, daysToAdd)
			if d.After(now) {
				break
			}
		}
		return d.Format("20060102"), nil
	default:
		return "", fmt.Errorf("incorrect format of the Repeat parameter")
	}
}
