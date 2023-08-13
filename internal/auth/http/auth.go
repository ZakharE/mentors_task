package http

import (
	"context"
	"mentoring/internal/auth/service"
	"net/http"

	"github.com/go-chi/chi"
)

type (
	Server struct {
		mux     *chi.Mux
		service *service.Auth
	}
)

func New(m *chi.Mux, service *service.Auth) {
	server := &Server{
		service: service,
	}

	m.Post("/api/auth", server.login)
}

func(s *Server) Start() {
	http.ListenAndServe(":8080", s.mux)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	accToken, refreshToken, err := s.service.IssueTokens(ctx, username, password)
	if err  != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	w.Header().Add("access_token", accToken)
	w.Header().Add("refresh_token", refreshToken)
}

func()
