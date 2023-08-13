package service

import (
	"context"
	"errors"
	"mentoring/internal/auth/models"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var secret = []byte("super_secret")

type (
	storage interface {
		Get(ctx context.Context, user models.User) error
	}

	Auth struct {
		storage storage
	}
)

func New(s storage) *Auth {
	return &Auth{storage: s}
}
//question: how to properly use context with cancel? 
func (a *Auth) IssueTokens(ctx context.Context, username, password string) (string, string, error) {
	u := models.User{
		Username: username,
		Password: password,
	}
	err := a.storage.Get(ctx, u)
	if errors.Is(err, models.ErrNotFound) {
		return "", "", err
	}
	//TODO save user's tokens to storage to mark them as inactive  in future   
	accessClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccToken, err := accessToken.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	return signedAccToken, signedRefreshToken, err
}

func (a *Auth) Verify() {
	panic("not impl")
}

func (a *Auth) Logout() {
	panic("not impl")
}
