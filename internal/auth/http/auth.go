package http

import (
	"context"
	"errors"
	"fmt"
	"mentoring/internal/auth/models"
	"mentoring/internal/auth/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"golang.org/x/exp/slog"
)

const (
	HeaderAccessToken  = "access_token"
	HeaderRefreshToken = "refresh_token"
)

type (
	Server struct {
		mux     *chi.Mux
		service *service.Auth
	}
)

func New(m *chi.Mux, service *service.Auth) *Server {
	server := &Server{
		service: service,
		mux:     m,
	}

	return server
}

func (s *Server) Start() {
	s.mux.Post("/api/auth", s.login)
	s.mux.Post("/api/verify", s.verify)
	s.mux.Post("/api/logout", s.logout)

	server := http.Server{Addr: ":8080", Handler: s.mux}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig

		shutdownContext, _ := context.WithTimeout(serverCtx, time.Second*30)
		go func() {
			<-shutdownContext.Done()
			if shutdownContext.Err() == context.DeadlineExceeded {
				os.Exit(1)
			}
		}()

		server.Shutdown(shutdownContext)
		serverStopCtx()
	}()

	slog.Info("Starting a server...")
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		os.Exit(1)
	}

	<-serverCtx.Done()
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	accToken, refreshToken, err := s.service.Auth(ctx, username, password)
	switch {
	case errors.Is(err, models.ErrNotFound):
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	case err != nil:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	s.setCookies(w, accToken, refreshToken)
}

func (s *Server) verify(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	accToken, err := r.Cookie(HeaderAccessToken)
	if err == http.ErrNoCookie {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	err = s.service.Verify(ctx, accToken.Value)
	switch {
	case errors.Is(err, models.ErrClaimsParse):
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	case errors.Is(err, models.ErrNotFound):
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	case errors.Is(err, models.ErrTokenInvalid):
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	//not sure what should exactly we do
	refreshToken, err := r.Cookie(HeaderRefreshToken)
	if err == http.ErrNoCookie {
		return
	}

	if err = s.service.Verify(ctx, refreshToken.Value); err != nil {
		return //not sure what should exactly we do
	}
	newAccToken, newRefreshToken, err := s.service.IssueTokensWithExpiration(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	s.service.Deactivate(ctx, accToken.Value)
	s.service.Deactivate(ctx, refreshToken.Value)
	s.setCookies(w, newAccToken, newRefreshToken)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	accessToken, err := r.Cookie(HeaderAccessToken)
	if err == http.ErrNoCookie {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	ctx := context.Background()
	if s.service.Verify(ctx, accessToken.Value) != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	s.service.Deactivate(ctx, accessToken.Value)

	refreshToken, err := r.Cookie(HeaderRefreshToken)

	if err == http.ErrNoCookie {
		return
	}
	s.service.Deactivate(ctx, refreshToken.Value)
}

func (s *Server) setCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	accCookie := http.Cookie{
		Name:    HeaderAccessToken,
		Value:   accessToken,
		Expires: time.Now().Add(5 * time.Minute),
	}
	refreshCookie := http.Cookie{
		Name:    HeaderRefreshToken,
		Value:   refreshToken,
		Expires: time.Now().Add(60 * time.Minute),
	}
	http.SetCookie(w, &accCookie) // question: set expiration to cookie and token?
	http.SetCookie(w, &refreshCookie)
}
