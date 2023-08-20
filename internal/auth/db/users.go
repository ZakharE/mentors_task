package db

import (
	"context"
	"mentoring/internal/auth/models"
	"sync"
)

type (
	Storage struct {
		entries sync.Map
	}
)

func New() *Storage {
	return &Storage{}
}

//question: maybe the better way to return error
func (s *Storage) HasEntry(ctx context.Context, user models.User) bool {
	_, ok:= s.entries.Load(user.Username)
	return ok
}
