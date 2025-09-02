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
		Status:      ifEmptyStatus(in.Status, StatusPending),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Create(t)
}

func (s *Service) Get(id int64) (Task, error) { return s.repo.GetByID(id) }
func (s *Service) Delete(id int64) error      { return s.repo.Delete(id) }
func (s *Service) List(st *Status, p, ps int) ([]Task, int, error) {
	return s.repo.List(st, p, ps)
}
func (s *Service) Update(id int64, in UpdateTaskInput) (Task, error) {
	t, err := s.repo.Update(id, in)
	if err != nil {
		return t, err
	}
	// Touch UpdatedAt
	now := time.Now().UTC()
	title, desc, status := t.Title, t.Description, t.Status
	t.UpdatedAt = now
	_, _ = s.repo.Update(id, UpdateTaskInput{Title: &title, Description: &desc, Status: &status})
	return t, nil
}

func ifEmptyStatus(given, fallback Status) Status {
	if given == "" {
		return fallback
	}
	return given
}
