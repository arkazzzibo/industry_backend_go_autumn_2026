package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{
		clock: clock,
		tasks: make(map[string]Task),
	}
}
func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	trimTitle := strings.TrimSpace(title)
	if trimTitle == "" {
		return Task{}, ErrInvalidTitle
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	id := strconv.FormatUint(r.seq, 10)
	task := Task{
		ID:        id,
		Title:     trimTitle,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}

	r.tasks[id] = task
	return task, nil
}
func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, false
	}
	return task, true
}
func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()

	result := make([]Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		if task.Done == done {
			result = append(result, task)
		}
	}

	r.mu.RUnlock()

	sort.Slice(result, func(i, j int) bool {
		if result[i].UpdatedAt.Equal(result[j].UpdatedAt) {
			return result[i].ID < result[j].ID
		}
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	return result
}
func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	if task.Done == done {
		return task, nil
	}

	task.Done = done
	task.UpdatedAt = r.clock.Now()
	r.tasks[id] = task

	return task, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &httpHandler{repo: repo}
}
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/tasks" {
		switch r.Method {
		case http.MethodGet:
			h.handleList(w, r)
		case http.MethodPost:
			h.handleCreate(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if strings.HasPrefix(r.URL.Path, "/tasks/") {
		id := strings.TrimPrefix(r.URL.Path, "/tasks/")

		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.handleGet(w, r, id)
		case http.MethodPatch:
			h.handlePatch(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	http.NotFound(w, r)
}
func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Title *string `json:"title"`
	}

	if err := decodeStrictJSON(r.Body, &request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.Title == nil {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	task, err := h.repo.Create(*request.Title)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}
func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)
	if !ok {
		http.Error(w, "не найдено", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, task)

}

func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	done := false
	values, ok := r.URL.Query()["done"]
	if ok {
		if len(values) != 1 {
			http.Error(w, "лишние значения", http.StatusBadRequest)
			return
		}

		switch values[0] {
		case "true":
			done = true
		case "false":
			done = false
		default:
			http.Error(w, "невалидный done", http.StatusBadRequest)
			return
		}
	}

	tasks := h.repo.List(done)
	if tasks == nil {
		tasks = []Task{}
	}
	writeJSON(w, http.StatusOK, tasks)
}
func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var request struct {
		Done *bool `json:"done"`
	}

	if err := decodeStrictJSON(r.Body, &request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.Done == nil {
		http.Error(w, "done is required", http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *request.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, task)
}
func decodeStrictJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields() // ошибка при неизвестных полях

	if err := dec.Decode(v); err != nil {
		return err
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain only one JSON object")
		}
		return err
	}

	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
