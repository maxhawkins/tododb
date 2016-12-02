package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type API struct {
	db *sqlx.DB
}

func (a *API) HandleListTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := ListTodosQuery{}.Exec(a.db)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	js, err := json.MarshalIndent(todos, "", "\t")
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(js)
}

func (a *API) HandleGetTodo(w http.ResponseWriter, r *http.Request) {
	idStr, _ := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)

	todo, err := GetTodoQuery{ID: id}.Exec(a.db)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	js, err := json.MarshalIndent(todo, "", "\t")
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(js)
}

func (a *API) HandleUpdateTodo(w http.ResponseWriter, r *http.Request) {
	idStr, _ := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := UpdateTodoCommand{todo}
	if err := cmd.Exec(id, a.db); err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	js, err := json.MarshalIndent(todo, "", "\t")
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(js)
}

func (a *API) HandleAddTodo(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := SaveTodoCommand{todo}
	updated, err := cmd.Exec(a.db)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	js, err := json.MarshalIndent(updated, "", "\t")
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(js)
}

func (a *API) HandleSaveList(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var js interface{}
	if err := json.Unmarshal(body, &js); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = a.db.Exec(
		"UPDATE lists SET data=$1 WHERE list_id = 0", body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintln(w, `{"status": "OK"}`)
}

func (a *API) HandleGetList(w http.ResponseWriter, r *http.Request) {
	row := a.db.QueryRow("SELECT data FROM lists WHERE list_id = 0")

	var data []byte
	if err := row.Scan(&data); err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}
