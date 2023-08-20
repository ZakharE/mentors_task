package main

import (
	"mentoring/internal/auth/db"
	"mentoring/internal/auth/http"
	"mentoring/internal/auth/service"

	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()
	storage := db.New()
	tokenStorage := db.NewTokenStrorage()
	s := service.New(storage, tokenStorage)
	server := http.New(r, s)
	server.Start()
}
