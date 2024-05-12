package main

import (
	"bytes"
	"encoding/json"
	"net/http"
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

	result, err := NextDate(dNow, date, repeat, false)

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

	if resultValidate := validateAndUpdateTask(task, false); resultValidate != nil {
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
	idInt, err := validateTaskID(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}

	task, err := store.getTask(idInt)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
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

	idInt, err := validateTaskID(task.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&Result{Error: err.Error()})
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

	if resultValidate := validateAndUpdateTask(task, false); resultValidate != nil {
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

func checkDoneTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	idInt, err := validateTaskID(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}
	task, err := store.getTask(idInt)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}

	if len(task.Repeat) == 0 {
		deleteTask(w, r)
		return
	}

	if resultValidate := validateAndUpdateTask(task, true); resultValidate != nil {
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

func deleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	idInt, err := validateTaskID(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}

	err = store.deleteTask(idInt)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(&Result{Error: err.Error()})
		w.Write(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string]any{})
	w.Write(res)
}
