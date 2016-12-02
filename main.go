package main

import (
	"flag"
	"fmt"
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
		dbPath = flag.String("db", "postgres://localhost/tododb?sslmode=disable", "db location")
		port   = flag.Int("port", 7766, "http port")
	)
	flag.Parse()

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

	m.PathPrefix("/").Handler(http.FileServer(http.Dir("www")))

	var handler http.Handler
	handler = handlers.LoggingHandler(os.Stderr, m)

	addr := fmt.Sprint(":", *port)
	fmt.Fprintln(os.Stderr, "listening at", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
