package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/alle-assignment/internal/platform/response"
	"github.com/alle-assignment/internal/task"
)

type Handler struct {
	svc *task.Service
}

func New(svc *task.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/tasks", h.tasks)
	mux.HandleFunc("/tasks/", h.taskByID)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) tasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTasksCursor(w, r)
	case http.MethodPost:
		h.createTask(w, r)
	default:
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var in task.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(in.Title) == "" {
		response.Error(w, http.StatusBadRequest, "title is required")
		return
	}
	if in.Status == "" {
		in.Status = task.StatusPending
	}
	t, err := h.svc.Create(in)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, t)
}

func (h *Handler) taskByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	switch r.Method {
	case http.MethodGet:
		t, err := h.svc.Get(id)
		if err != nil {
			if errors.Is(err, task.ErrNotFound) {
				response.Error(w, http.StatusNotFound, "not found")
				return
			}
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		response.JSON(w, http.StatusOK, t)
	case http.MethodPut:
		var in task.UpdateTaskInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			response.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		t, err := h.svc.Update(id, in)
		if err != nil {
			if errors.Is(err, task.ErrNotFound) {
				response.Error(w, http.StatusNotFound, "not found")
				return
			}
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		response.JSON(w, http.StatusOK, t)
	case http.MethodDelete:
		if err := h.svc.Delete(id); err != nil {
			if errors.Is(err, task.ErrNotFound) {
				response.Error(w, http.StatusNotFound, "not found")
				return
			}
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) listTasksCursor(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	var status *task.Status
	if s := q.Get("status"); s != "" {
		tmp := task.Status(s)
		status = &tmp
	}
	cursorStr := q.Get("cursor")
	after, err := decodeCursor(cursorStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid cursor")
		return
	}
	items, next, err := h.svc.ListWithCursor(after, limit, status)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	nextStr, err := encodeCursor(next)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to encode cursor")
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"data":        items,
		"next_cursor": nextStr,
		"limit":       limit,
	})
}

func decodeCursor(s string) (*task.Cursor, error) {
	if s == "" {
		return nil, nil
	}
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var c task.Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func encodeCursor(c *task.Cursor) (string, error) {
	if c == nil {
		return "", nil
	}
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
