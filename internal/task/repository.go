package task

import (
	"errors"
	"sort"
	"sync"
)

var ErrNotFound = errors.New("task not found")

// Repository abstracts persistence.
type Repository interface {
	Create(t Task) (Task, error)
	GetByID(id int64) (Task, error)
	Update(id int64, in UpdateTaskInput) (Task, error)
	Delete(id int64) error
	// List supports optional status filter + pagination.
	List(status *Status, page, pageSize int) ([]Task, int, error)
}

type InMemoryRepository struct {
	mu    sync.RWMutex
	seq   int64
	items map[int64]Task
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{items: make(map[int64]Task)}
}

func (r *InMemoryRepository) Create(t Task) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	t.ID = r.seq
	r.items[t.ID] = t
	return t, nil
}

func (r *InMemoryRepository) GetByID(id int64) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.items[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return t, nil
}

func (r *InMemoryRepository) Update(id int64, in UpdateTaskInput) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.items[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	if in.Title != nil {
		t.Title = *in.Title
	}
	if in.Description != nil {
		t.Description = *in.Description
	}
	if in.Status != nil {
		t.Status = *in.Status
	}
	r.items[id] = t
	return t, nil
}

func (r *InMemoryRepository) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *InMemoryRepository) List(status *Status, page, pageSize int) ([]Task, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []Task
	for _, t := range r.items {
		if status != nil && t.Status != *status {
			continue
		}
		all = append(all, t)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

	total := len(all)
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}

	start := (page - 1) * pageSize
	if start >= total {
		return []Task{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}
