package db

import (
	"context"
	"sync"
)

type (
	TokenStorage struct {
		entries sync.Map
	}
)

func NewTokenStrorage() *TokenStorage {
	return &TokenStorage{}
}

// should i pass context here? Probably, yes. Since interface of storage can be applied to real db in future
func (s *TokenStorage) Save(ctx context.Context, token string) {
	s.entries.Store(token, true)
}

func (s *TokenStorage) Get(ctx context.Context, token string) bool {
	_, ok := s.entries.Load(token)
	return ok
}

func (s *TokenStorage) Delete(ctx context.Context, token string) {
	s.entries.Delete(token)
}
