package main

import (
	"log"
	"net/http"

	"github.com/alle-assignment/internal/httpapi"
	"github.com/alle-assignment/internal/task"
)

func main() {
	repo := task.NewInMemoryRepository()
	svc := task.NewService(repo)
	handler := httpapi.New(svc)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	const addr = ":8080"
	log.Printf("Task Management Service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
