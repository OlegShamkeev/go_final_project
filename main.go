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
	case "w":
		if len(arrayParams) <= 1 {
			return "", fmt.Errorf("incorrect format of the repeat parameter")
		}
		daysMap := make(map[int]int)
		days := strings.Split(arrayParams[1], ",")
		for _, i := range days {
			day, err := strconv.Atoi(i)
			if day > 7 || err != nil {
				return "", fmt.Errorf("incorrect format of the repeat parameter")
			}
			if day == 7 {
				day = 0
			}
			daysMap[day]++
		}
		for {
			d = d.AddDate(0, 0, 1)
			if _, ok := daysMap[int(d.Weekday())]; d.After(now) && ok {
				break
			}
		}
		return d.Format("20060102"), nil
	case "m":
		if len(arrayParams) <= 1 {
			return "", fmt.Errorf("incorrect format of the repeat parameter")
		}

		daysMap := make(map[int]int)
		days := strings.Split(arrayParams[1], ",")
		for _, i := range days {
			day, err := strconv.Atoi(i)
			if day > 31 || day < -2 || err != nil {
				return "", fmt.Errorf("incorrect format of the repeat parameter")
			}
			daysMap[day]++
		}

		monthsMap := make(map[int]int)
		if len(arrayParams) > 2 {
			months := strings.Split(arrayParams[2], ",")
			for _, i := range months {
				month, err := strconv.Atoi(i)
				if month > 12 || err != nil {
					return "", fmt.Errorf("incorrect format of the repeat parameter")
				}
				monthsMap[month]++
			}
		} else {
			for i := 1; i < 13; i++ {
				monthsMap[i]++
			}
		}

		for {
			d = d.AddDate(0, 0, 1)

			_, ok1 := daysMap[d.Day()]

			t := time.Date(d.Year(), d.Month(), 32, 0, 0, 0, 0, time.UTC)
			daysInMonth := 32 - t.Day()

			backwardKey := d.Day() - daysInMonth - 1
			_, ok2 := daysMap[backwardKey]

			if _, ok3 := monthsMap[int(d.Month())]; (ok1 || ok2) && d.After(now) && ok3 {
				break
			}
		}
		return d.Format("20060102"), nil
	default:
		return "", fmt.Errorf("incorrect format of the Repeat parameter")
	}
}
