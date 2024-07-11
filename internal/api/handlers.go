package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/OlegShamkeev/go_final_project/internal/config"
	"github.com/OlegShamkeev/go_final_project/internal/nextdate"
	"github.com/OlegShamkeev/go_final_project/internal/storage"
	"github.com/OlegShamkeev/go_final_project/internal/task"

	"github.com/golang-jwt/jwt"
)

var cfg *config.Config
var store *storage.Storage
var secret []byte

const secretLength = 20

type Result struct {
	Id    int    `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func NewApi(config *config.Config, strg *storage.Storage) {
	cfg = config
	store = strg
	if len(cfg.Password) > 0 {
		secret = generateSecret()
	}
}

func generateSecret() []byte {
	rnd := rand.NewSource(time.Now().Unix())
	result := make([]byte, 0, secretLength)
	for i := 0; i < secretLength; i++ {
		randomNumber := rnd.Int63()
		result = append(result, byte(randomNumber%26+97))
	}
	return result
}

func validateTaskID(id string) (int, error) {
	if len(id) == 0 {
		return 0, fmt.Errorf("no id parameter")
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}
	return idInt, nil
}

func GetNextDate(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application-json")

	dNow, err := time.Parse("20060102", r.URL.Query().Get("now"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	result, err := nextdate.NextDate(dNow, date, repeat, false)

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

func errorMessage(w http.ResponseWriter, status uint, msg any) {
	writeJson(w, status,
		&Result{Error: fmt.Sprint(msg)},
	)
}

func writeJson(w http.ResponseWriter, status uint, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(int(status))

	res, _ := json.Marshal(data)
	_, err := w.Write(res)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
	}
}

func PostTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var task *task.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		errorMessage(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
		errorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	if resultValidate := task.ValidateAndUpdateTask(false); resultValidate != "" {
		errorMessage(w, http.StatusBadRequest, resultValidate)
		return
	}

	id, err := store.CreateTask(task)

	if err != nil {
		errorMessage(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	res, _ := json.Marshal(&Result{Id: id})
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		return
	}
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	search := r.URL.Query().Get("search")

	tasks, err := store.GetTasks(search)
	if err != nil {
		errorMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string][]task.Task{"tasks": tasks})
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		return
	}
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	idInt, err := validateTaskID(id)

	if err != nil {
		errorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := store.GetTask(idInt)

	if err != nil {
		errorMessage(w, http.StatusNotFound, err.Error())
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
	var task *task.Task

	_, err := buf.ReadFrom(r.Body)

	if err != nil {
		errorMessage(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &task); err != nil {
		errorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	idInt, err := validateTaskID(task.Id)
	if err != nil {
		errorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	_, err = store.GetTask(idInt)

	if err != nil {
		errorMessage(w, http.StatusNotFound, err.Error())
		return
	}

	if resultValidate := task.ValidateAndUpdateTask(false); resultValidate != "" {
		errorMessage(w, http.StatusNotFound, resultValidate)
		return
	}
	err = store.UpdateTask(task)
	if err != nil {
		errorMessage(w, http.StatusNotFound, err.Error())
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
	idInt, err := validateTaskID(id)

	if err != nil {
		errorMessage(w, http.StatusBadRequest, err.Error())
		return
	}
	task, err := store.GetTask(idInt)

	if err != nil {
		errorMessage(w, http.StatusNotFound, err.Error())
		return
	}

	if len(task.Repeat) == 0 {
		DeleteTask(w, r)
		return
	}

	if resultValidate := task.ValidateAndUpdateTask(true); resultValidate != "" {
		errorMessage(w, http.StatusBadRequest, resultValidate)
		return
	}

	err = store.UpdateTask(task)

	if err != nil {
		errorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string]any{})
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		return
	}
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	idInt, err := validateTaskID(id)

	if err != nil {
		errorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	err = store.DeleteTask(idInt)

	if err != nil {
		errorMessage(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string]any{})
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		return
	}
}

func AuthAndGenerateToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		errorMessage(w, http.StatusBadRequest, err.Error())
		return
	}
	p := map[string]string{}
	if err := json.Unmarshal(buf.Bytes(), &p); err != nil {
		errorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	if p["password"] != cfg.Password {
		errorMessage(w, http.StatusUnauthorized, "wrong password")
		return
	}

	result := sha256.Sum256([]byte(cfg.Password))
	claims := jwt.MapClaims{
		"hashPass": hex.EncodeToString(result[:]),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		errorMessage(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(&map[string]string{"token": signedToken})
	_, err = w.Write(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during writing data to response writer %s", err.Error())
		return
	}
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(cfg.Password) > 0 {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")

			var token string
			cookie, err := r.Cookie("token")
			if err == nil {
				token = cookie.Value
			}

			jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
				return secret, nil
			})

			if err != nil {
				errorMessage(w, http.StatusUnauthorized, err.Error())
				return
			}
			if !jwtToken.Valid {
				errorMessage(w, http.StatusUnauthorized, "jwt token isn't valid")
				return
			}

			res, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok {
				errorMessage(w, http.StatusUnauthorized, "failed to typecast to jwt.MapCalims")
				return
			}

			hashPassRaw := res["hashPass"]
			hashPass, ok := hashPassRaw.(string)
			if !ok {
				errorMessage(w, http.StatusUnauthorized, "failed to typecase password hash to string")
				return
			}

			result := sha256.Sum256([]byte(cfg.Password))
			if hashPass != hex.EncodeToString(result[:]) {
				errorMessage(w, http.StatusUnauthorized, "token password hash doesn't match")
				return
			}
		}
		next(w, r)
	})
}
