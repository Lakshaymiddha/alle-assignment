package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/alle-assignment/internal/task"
	"github.com/stretchr/testify/assert"
)

func setupTestMux() *http.ServeMux {
	repo := task.NewInMemoryRepository()
	svc := task.NewService(repo)
	h := New(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func TestHandler_CreateTask(t *testing.T) {
	mux := setupTestMux()

	reqBody := `{"title":"Test","description":"desc","status":"Pending"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Test", resp["title"])
}

func TestHandler_CreateTask_BadRequest(t *testing.T) {
	mux := setupTestMux()

	// Missing title
	reqBody := `{"description":"desc","status":"Pending"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetTask_NotFound(t *testing.T) {
	mux := setupTestMux()

	req := httptest.NewRequest("GET", "/tasks/9999", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_GetTask_Success(t *testing.T) {
	mux := setupTestMux()

	// First, create a task
	reqBody := `{"title":"TestGet","description":"desc","status":"Pending"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	id := int(resp["id"].(float64))

	// Now, get the task
	req = httptest.NewRequest("GET", "/tasks/"+strconv.Itoa(id), nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var getResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &getResp)
	assert.Equal(t, "TestGet", getResp["title"])
}

func TestHandler_DeleteTask(t *testing.T) {
	mux := setupTestMux()

	// Create a task
	reqBody := `{"title":"TestDelete","description":"desc","status":"Pending"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	id := int(resp["id"].(float64))

	// Delete the task
	req = httptest.NewRequest("DELETE", "/tasks/"+strconv.Itoa(id), nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Try to get the deleted task
	req = httptest.NewRequest("GET", "/tasks/"+strconv.Itoa(id), nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_UpdateTask(t *testing.T) {
	mux := setupTestMux()

	// Create a task
	reqBody := `{"title":"TestUpdate","description":"desc","status":"Pending"}`
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	id := int(resp["id"].(float64))

	// Update the task
	updateBody := `{"title":"UpdatedTitle"}`
	req = httptest.NewRequest("PUT", "/tasks/"+strconv.Itoa(id), bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var updateResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &updateResp)
	assert.Equal(t, "UpdatedTitle", updateResp["title"])
}

func TestHandler_ListTasks(t *testing.T) {
	mux := setupTestMux()

	// Create a few tasks
	for i := 0; i < 3; i++ {
		reqBody := `{"title":"Task` + strconv.Itoa(i) + `","description":"desc","status":"Pending"}`
		req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}

	// List tasks
	req := httptest.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var listResp []map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &listResp)
	assert.GreaterOrEqual(t, len(listResp), 3)
}

func TestHandler_GetTask_InvalidID(t *testing.T) {
	mux := setupTestMux()

	// Non-numeric ID
	req := httptest.NewRequest("GET", "/tasks/abc", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetTask_ZeroID(t *testing.T) {
	mux := setupTestMux()

	// Zero ID (if your handler treats zero as invalid)
	req := httptest.NewRequest("GET", "/tasks/0", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
