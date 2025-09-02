package main

import (
	"log"
	"net/http"

	"github.com/task-management-service/internal/httpapi"
	"github.com/task-management-service/internal/task"
)

func main() {
	repo := task.NewInMemoryRepository()
	svc := task.NewService(repo)
	h := httpapi.New(svc)

	mux := http.NewServeMux()
	h.Register(mux)

	addr := ":8080"
	log.Printf("task-svc listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
