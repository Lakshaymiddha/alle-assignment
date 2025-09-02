package task

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "time"
)

func TestInMemoryRepository_CRUD(t *testing.T) {
    repo := NewInMemoryRepository()

    // Create
    task := Task{Title: "Test", Description: "desc", Status: StatusPending}
    created, err := repo.Create(task)
    assert.NoError(t, err)
    assert.NotZero(t, created.ID)

    // GetByID
    got, err := repo.GetByID(created.ID)
    assert.NoError(t, err)
    assert.Equal(t, created.Title, got.Title)

    // Update
    newTitle := "Updated"
    input := UpdateTaskInput{Title: &newTitle}
    updated, err := repo.Update(created.ID, input)
    assert.NoError(t, err)
    assert.Equal(t, newTitle, updated.Title)

    // Delete
    err = repo.Delete(created.ID)
    assert.NoError(t, err)
    _, err = repo.GetByID(created.ID)
    assert.Error(t, err)
}

func TestInMemoryRepository_ListByCursor(t *testing.T) {
    repo := NewInMemoryRepository()
    for i := 0; i < 5; i++ {
        repo.Create(Task{Title: "T", Status: StatusPending, CreatedAt: time.Now().Add(time.Duration(i) * time.Second)})
    }
    tasks, next, err := repo.ListByCursor(nil, 3, nil)
    assert.NoError(t, err)
    assert.Len(t, tasks, 3)
    assert.NotNil(t, next)
}