package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/task-management-service/internal/platform/response"
	"github.com/task-management-service/internal/task"
)

type Handler struct{ svc *task.Service }

func New(s *task.Service) *Handler { return &Handler{svc: s} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /tasks", h.listTasks)
	mux.HandleFunc("POST /tasks", h.createTask)
	mux.HandleFunc("GET /tasks/{id}", h.getTask)
	mux.HandleFunc("PUT /tasks/{id}", h.updateTask)
	mux.HandleFunc("DELETE /tasks/{id}", h.deleteTask)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var in task.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Title) == "" {
		response.Error(w, http.StatusBadRequest, "invalid input: title required")
		return
	}
	t, err := h.svc.Create(in)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, t)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	tk, err := h.svc.Get(id)
	if err != nil {
		if errors.Is(err, task.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, tk)
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var in task.UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	tk, err := h.svc.Update(id, in)
	if err != nil {
		if errors.Is(err, task.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, tk)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, task.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusNoContent, nil)
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))

	var st *task.Status
	if s := q.Get("status"); s != "" {
		tmp := task.Status(s)
		st = &tmp
	}

	items, total, err := h.svc.List(st, page, pageSize)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"data":       items,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + pageSize - 1) / pageSize,
	})
}
