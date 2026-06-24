// Package handlers contains HTTP handlers for the Todo API.
// Handlers depend on the repository.TodoRepository interface, not on
// MongoDB directly, keeping the transport layer decoupled from storage.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/example/todo-go-app/internal/models"
	"github.com/example/todo-go-app/internal/repository"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const requestTimeout = 5 * time.Second

// TodoHandler groups HTTP handlers for the /todos resource.
type TodoHandler struct {
	repo repository.TodoRepository
}

// NewTodoHandler builds a TodoHandler backed by the given repository.
func NewTodoHandler(repo repository.TodoRepository) *TodoHandler {
	return &TodoHandler{repo: repo}
}

// Create handles POST /todos
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Title == "" {
		respondError(w, http.StatusBadRequest, "title is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	todo := &models.Todo{Title: body.Title}
	if err := h.repo.Create(ctx, todo); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create todo")
		return
	}

	respondJSON(w, http.StatusCreated, todo)
}

// List handles GET /todos
func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	todos, err := h.repo.List(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch todos")
		return
	}

	respondJSON(w, http.StatusOK, todos)
}

// Get handles GET /todos/{id}
func (h *TodoHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := objectIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid todo id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	todo, err := h.repo.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		respondError(w, http.StatusNotFound, "todo not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch todo")
		return
	}

	respondJSON(w, http.StatusOK, todo)
}

// Update handles PUT /todos/{id}
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := objectIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid todo id")
		return
	}

	var update models.TodoUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if update.Title == nil && update.Done == nil {
		respondError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	todo, err := h.repo.Update(ctx, id, update)
	if errors.Is(err, repository.ErrNotFound) {
		respondError(w, http.StatusNotFound, "todo not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update todo")
		return
	}

	respondJSON(w, http.StatusOK, todo)
}

// Delete handles DELETE /todos/{id}
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := objectIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid todo id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	if err := h.repo.Delete(ctx, id); errors.Is(err, repository.ErrNotFound) {
		respondError(w, http.StatusNotFound, "todo not found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete todo")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func objectIDFromRequest(r *http.Request) (primitive.ObjectID, error) {
	idParam := mux.Vars(r)["id"]
	return primitive.ObjectIDFromHex(idParam)
}
