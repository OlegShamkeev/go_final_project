package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func getNextDate(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application-json")

	dNow, err := time.Parse("20060102", r.URL.Query().Get("now"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	result, err := NextDate(dNow, date, repeat)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b := []byte(result)

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		return
	}
}

func postTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var task Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}

	if len(strings.TrimSpace(task.Title)) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&Result{Error: "field title couldn't be empty"})
		w.Write(res)
		return
	}

	if len(strings.TrimSpace(task.Date)) == 0 {
		task.Date = time.Now().Format("20060102")
	} else {
		dateParsed, err := time.Parse("20060102", task.Date)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			res, _ := json.Marshal(&Result{Error: err.Error()})
			w.Write(res)
			return
		}
		if len(strings.TrimSpace(task.Repeat)) > 0 {
			task.Date, err = NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				res, _ := json.Marshal(&Result{Error: err.Error()})
				w.Write(res)
				return
			}
		} else if dateParsed.Before(time.Now()) {
			task.Date = time.Now().Format("20060102")
		}
	}
	id, err := store.createTask(task)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}
	w.WriteHeader(http.StatusCreated)
	res, _ := json.Marshal(&Result{Id: id})
	w.Write(res)
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	search := r.URL.Query().Get("search")

	tasks, err := store.getTasks(search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string][]Task{"tasks": tasks})
	w.Write(res)
}
