package db

import (
	"context"
	"mentoring/internal/auth/models"
)

type (
	Storage struct {
	}
)

func New() *Storage {
	return &Storage{}
}
func (s *Storage) Get(ctx context.Context, user models.User) error {

	return nil
}

