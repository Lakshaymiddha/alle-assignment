package task

import "time"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(in CreateTaskInput) (Task, error) {
	now := time.Now().UTC()
	t := Task{
		Title:       in.Title,
		Description: in.Description,
		Status:      in.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if t.Status == "" {
		t.Status = StatusPending
	}
	return s.repo.Create(t)
}

func (s *Service) Get(id int64) (Task, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Update(id int64, in UpdateTaskInput) (Task, error) {
	return s.repo.Update(id, in)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}

func (s *Service) ListWithCursor(after *Cursor, limit int, status *Status) ([]Task, *Cursor, error) {
	return s.repo.ListByCursor(after, limit, status)
}
