package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/todo-go-app/internal/models"
	"github.com/example/todo-go-app/internal/repository"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// fakeRepo is an in-memory repository.TodoRepository used for testing
// handlers without a real MongoDB instance.
type fakeRepo struct {
	todos map[primitive.ObjectID]models.Todo
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{todos: make(map[primitive.ObjectID]models.Todo)}
}

func (f *fakeRepo) Create(_ context.Context, todo *models.Todo) error {
	todo.ID = primitive.NewObjectID()
	f.todos[todo.ID] = *todo
	return nil
}

func (f *fakeRepo) List(_ context.Context) ([]models.Todo, error) {
	var out []models.Todo
	for _, t := range f.todos {
		out = append(out, t)
	}
	return out, nil
}

func (f *fakeRepo) Get(_ context.Context, id primitive.ObjectID) (*models.Todo, error) {
	t, ok := f.todos[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return &t, nil
}

func (f *fakeRepo) Update(_ context.Context, id primitive.ObjectID, update models.TodoUpdate) (*models.Todo, error) {
	t, ok := f.todos[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	if update.Title != nil {
		t.Title = *update.Title
	}
	if update.Done != nil {
		t.Done = *update.Done
	}
	f.todos[id] = t
	return &t, nil
}

func (f *fakeRepo) Delete(_ context.Context, id primitive.ObjectID) error {
	if _, ok := f.todos[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.todos, id)
	return nil
}

func TestCreate_Success(t *testing.T) {
	repo := newFakeRepo()
	h := NewTodoHandler(repo)

	body := bytes.NewBufferString(`{"title":"write tests"}`)
	req := httptest.NewRequest(http.MethodPost, "/todos", body)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var got models.Todo
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.Title != "write tests" {
		t.Errorf("expected title 'write tests', got %q", got.Title)
	}
}

func TestCreate_MissingTitle(t *testing.T) {
	repo := newFakeRepo()
	h := NewTodoHandler(repo)

	body := bytes.NewBufferString(`{"title":""}`)
	req := httptest.NewRequest(http.MethodPost, "/todos", body)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestGet_NotFound(t *testing.T) {
	repo := newFakeRepo()
	h := NewTodoHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/todos/"+primitive.NewObjectID().Hex(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": primitive.NewObjectID().Hex()})
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestUpdate_TogglesDone(t *testing.T) {
	repo := newFakeRepo()
	h := NewTodoHandler(repo)

	todo := &models.Todo{Title: "ship feature"}
	_ = repo.Create(context.Background(), todo)

	body := bytes.NewBufferString(`{"done":true}`)
	req := httptest.NewRequest(http.MethodPut, "/todos/"+todo.ID.Hex(), body)
	req = mux.SetURLVars(req, map[string]string{"id": todo.ID.Hex()})
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got models.Todo
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if !got.Done {
		t.Errorf("expected done=true, got false")
	}
}

func TestDelete_Success(t *testing.T) {
	repo := newFakeRepo()
	h := NewTodoHandler(repo)

	todo := &models.Todo{Title: "temp"}
	_ = repo.Create(context.Background(), todo)

	req := httptest.NewRequest(http.MethodDelete, "/todos/"+todo.ID.Hex(), nil)
	req = mux.SetURLVars(req, map[string]string{"id": todo.ID.Hex()})
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}
