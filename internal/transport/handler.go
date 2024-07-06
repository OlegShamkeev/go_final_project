package transport

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/OlegShamkeev/go_final_project/internal/config"
	"github.com/OlegShamkeev/go_final_project/internal/database"
	"github.com/OlegShamkeev/go_final_project/internal/services"
	"github.com/OlegShamkeev/go_final_project/internal/types"
	"github.com/OlegShamkeev/go_final_project/internal/utils"

	"github.com/golang-jwt/jwt"
)

var Cfg *config.Config
var Store *database.Storage
var Secret []byte

func GetNextDate(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application-json")

	dNow, err := time.Parse("20060102", r.URL.Query().Get("now"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	result, err := services.NextDate(dNow, date, repeat, false)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b := []byte(result)

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func PostTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var task *services.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}

	if resultValidate := task.ValidateAndUpdateTask(false); resultValidate != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(resultValidate)
		w.Write(res)
		return
	}

	id, err := Store.CreateTask(task)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}
	w.WriteHeader(http.StatusCreated)
	res, _ := json.Marshal(&types.Result{Id: id})
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	search := r.URL.Query().Get("search")

	tasks, err := Store.GetTasks(search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string][]services.Task{"tasks": tasks})
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	idInt, err := utils.ValidateTaskID(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}

	task, err := Store.GetTask(idInt)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(task)
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var buf bytes.Buffer
	var task *services.Task

	_, err := buf.ReadFrom(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}

	idInt, err := utils.ValidateTaskID(task.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}

	_, err = Store.GetTask(idInt)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}

	if resultValidate := task.ValidateAndUpdateTask(false); resultValidate != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(resultValidate)
		w.Write(res)
		return
	}
	err = Store.UpdateTask(task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string]any{})
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func CheckDoneTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	idInt, err := utils.ValidateTaskID(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}
	task, err := Store.GetTask(idInt)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}

	if len(task.Repeat) == 0 {
		DeleteTask(w, r)
		return
	}

	if resultValidate := task.ValidateAndUpdateTask(true); resultValidate != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(resultValidate)
		w.Write(res)
		return
	}

	err = Store.UpdateTask(task)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string]any{})
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	idInt, err := utils.ValidateTaskID(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}

	err = Store.DeleteTask(idInt)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string]any{})
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func AuthAndGenerateToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}
	p := map[string]string{}
	if err := json.Unmarshal(buf.Bytes(), &p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}

	if p["password"] != Cfg.Password {
		w.WriteHeader(http.StatusUnauthorized)
		res, _ := json.Marshal(&types.Result{Error: "wrong password"})
		w.Write(res)
		return
	}

	result := sha256.Sum256([]byte(Cfg.Password))
	claims := jwt.MapClaims{
		"hashPass": hex.EncodeToString(result[:]),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := jwtToken.SignedString(Secret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(&types.Result{Error: err.Error()})
		w.Write(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string]string{"token": signedToken})
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(Cfg.Password) > 0 {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")

			var token string
			cookie, err := r.Cookie("token")
			if err == nil {
				token = cookie.Value
			}

			jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
				return Secret, nil
			})

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				res, _ := json.Marshal(&types.Result{Error: err.Error()})
				w.Write(res)
				return
			}
			if !jwtToken.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				res, _ := json.Marshal(&types.Result{Error: "jwt token isn't valid"})
				w.Write(res)
				return
			}

			res, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				res, _ := json.Marshal(&types.Result{Error: "failed to typecast to jwt.MapCalims"})
				w.Write(res)
				return
			}

			hashPassRaw := res["hashPass"]
			hashPass, ok := hashPassRaw.(string)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				res, _ := json.Marshal(&types.Result{Error: "failed to typecase password hash to string"})
				w.Write(res)
				return
			}

			result := sha256.Sum256([]byte(Cfg.Password))
			if hashPass != hex.EncodeToString(result[:]) {
				w.WriteHeader(http.StatusUnauthorized)
				res, _ := json.Marshal(&types.Result{Error: "token password hash doesn't match"})
				w.Write(res)
				return
			}
		}
		next(w, r)
	})
}
