package task

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Cursor struct {
	Time time.Time `json:"t"`
	ID   int64     `json:"id"`
}

type Repository interface {
	Create(t Task) (Task, error)
	GetByID(id int64) (Task, error)
	Update(id int64, in UpdateTaskInput) (Task, error)
	Delete(id int64) error
	ListByCursor(after *Cursor, limit int, status *Status) ([]Task, *Cursor, error)
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
	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = t.CreatedAt
	}
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
	t.UpdatedAt = time.Now().UTC()
	r.items[id] = t
	return t, nil
}

func (r *InMemoryRepository) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[id]; !exists {
		return ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *InMemoryRepository) ListByCursor(after *Cursor, limit int, status *Status) ([]Task, *Cursor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []Task
	for _, t := range r.items {
		if status != nil && t.Status != *status {
			continue
		}
		if after != nil {
			if t.CreatedAt.Before(after.Time) || (t.CreatedAt.Equal(after.Time) && t.ID <= after.ID) {
				continue
			}
		}
		out = append(out, t)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})

	if limit <= 0 {
		limit = 10
	}
	if limit > len(out) {
		limit = len(out)
	}
	result := out[:limit]

	var next *Cursor
	if len(out) > limit {
		last := result[len(result)-1]
		next = &Cursor{Time: last.CreatedAt, ID: last.ID}
	}

	return result, next, nil
}

// Compile-time check that InMemoryRepository implements Repository interface
var _ Repository = (*InMemoryRepository)(nil)
