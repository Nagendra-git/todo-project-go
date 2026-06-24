// Package router wires HTTP routes to their handlers and middleware.
package router

import (
	"net/http"

	"github.com/example/todo-go-app/internal/handlers"
	"github.com/example/todo-go-app/internal/middleware"
	"github.com/gorilla/mux"
)

// New builds the application's HTTP handler, including middleware.
func New(todoHandler *handlers.TodoHandler, allowedOrigin string) http.Handler {
	r := mux.NewRouter()

	// Keep internal logging inside the router
	r.Use(middleware.Logging)

	r.HandleFunc("/healthz", healthCheck).Methods(http.MethodGet)

	todos := r.PathPrefix("/todos").Subrouter()
	todos.HandleFunc("", todoHandler.Create).Methods(http.MethodPost)
	todos.HandleFunc("", todoHandler.List).Methods(http.MethodGet)
	todos.HandleFunc("/{id}", todoHandler.Get).Methods(http.MethodGet)
	todos.HandleFunc("/{id}", todoHandler.Update).Methods(http.MethodPut)
	todos.HandleFunc("/{id}", todoHandler.Delete).Methods(http.MethodDelete)

	// Wrap the entire router with your CORS middleware.
	// This lets your middleware handle OPTIONS requests before Gorilla Mux filters by method.
	return middleware.CORS(allowedOrigin)(r)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
