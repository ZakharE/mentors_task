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
	s:= service.New(storage)
	http.New(r, s)
}
