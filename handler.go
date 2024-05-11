package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

	var task *Task
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

	if resultValidate := validateTask(task); resultValidate != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(resultValidate)
		w.Write(res)
		return
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

func getTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&Result{Error: "no id parameter"})
		w.Write(res)
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&Result{Error: fmt.Sprintf("field id should be number value: %s", err.Error())})
		w.Write(res)
		return
	}

	task, err := store.getTask(idInt)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(task)
	w.Write(res)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var buf bytes.Buffer
	var task *Task

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

	if len(task.Id) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&Result{Error: "no id parameter"})
		w.Write(res)
		return
	}
	idInt, err := strconv.Atoi(task.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&Result{Error: fmt.Sprintf("field id should be number value: %s", err.Error())})
		w.Write(res)
		return
	}

	_, err = store.getTask(idInt)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}

	if resultValidate := validateTask(task); resultValidate != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(resultValidate)
		w.Write(res)
		return
	}
	err = store.updateTask(task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string]any{})
	w.Write(res)
}
