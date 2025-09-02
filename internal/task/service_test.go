package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_CreateAndGetTask(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)

	input := CreateTaskInput{
		Title:       "Test Task",
		Description: "A test task",
		Status:      StatusPending,
	}
	created, err := svc.Create(input)
	assert.NoError(t, err)
	assert.Equal(t, input.Title, created.Title)

	got, err := svc.Get(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

func TestService_UpdateAndDeleteTask(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)

	input := CreateTaskInput{
		Title:       "Test Task",
		Description: "A test task",
		Status:      StatusPending,
	}
	created, _ := svc.Create(input)

	newTitle := "Updated Title"
	updateInput := UpdateTaskInput{Title: &newTitle}
	updated, err := svc.Update(created.ID, updateInput)
	assert.NoError(t, err)
	assert.Equal(t, newTitle, updated.Title)

	err = svc.Delete(created.ID)
	assert.NoError(t, err)
	_, err = svc.Get(created.ID)
	assert.Error(t, err)
}

func TestService_ListByCursor(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)

	for i := 0; i < 5; i++ {
		svc.Create(CreateTaskInput{
			Title:       "Task",
			Description: "desc",
			Status:      StatusPending,
		})
	}
	tasks, next, err := svc.ListWithCursor(nil, 3, nil)
	assert.NoError(t, err)
	assert.Len(t, tasks, 3)
	assert.NotNil(t, next)
}
