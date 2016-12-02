package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

func main() {
	var (
		dbPath = flag.String("db", "postgres://postgres/tododb?sslmode=disable", "postgres db url")
		port   = flag.Int("port", 7766, "http port")
		apiURL = flag.String("api", "", "api url")
	)
	flag.Parse()

	indexHTML, err := ioutil.ReadFile("www/index.html")
	if err != nil {
		log.Fatal("couldn't find index.html")
	}
	indexHTML = bytes.Replace(indexHTML, []byte("{{API_URL}}"), []byte(*apiURL), -1)

	db := sqlx.MustConnect("postgres", *dbPath)
	defer db.Close()

	if err := SetupTables(db); err != nil {
		log.Fatal(err)
	}

	api := &API{db}

	m := mux.NewRouter()
	m.HandleFunc("/todos/:id", api.HandleGetTodo).Methods("GET")
	m.HandleFunc("/todos/:id", api.HandleUpdateTodo).Methods("PUT")
	m.HandleFunc("/todos", api.HandleAddTodo).Methods("POST")
	m.HandleFunc("/todos", api.HandleListTodos).Methods("GET")

	m.HandleFunc("/lists/1", api.HandleGetList).Methods("GET")
	m.HandleFunc("/lists/1", api.HandleSaveList).Methods("PUT")

	static := http.FileServer(http.Dir("www"))

	m.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(indexHTML)
			return
		}

		static.ServeHTTP(w, r)
	})

	var handler http.Handler
	handler = handlers.LoggingHandler(os.Stderr, m)

	addr := fmt.Sprint(":", *port)
	fmt.Fprintln(os.Stderr, "listening at", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
