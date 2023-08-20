package service

import (
	"context"
	"fmt"
	"mentoring/internal/auth/models"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var secret = []byte("super_secret")

type (
	storage interface {
		HasEntry(ctx context.Context, user models.User) bool
	}
	tokenStorage interface {
		Get(ctx context.Context, token string) bool
		Save(ctx context.Context, token string)
		Delete(ctx context.Context, token string)
	}

	Auth struct {
		storage      storage
		tokenStorage tokenStorage
	}
)

func New(s storage, t tokenStorage) *Auth {
	return &Auth{
		storage:      s,
		tokenStorage: t,
	}
}

// question: how to properly use context with cancel?
func (a *Auth) Auth(ctx context.Context, username, password string) (string, string, error) {
	u := models.User{
		Username: username,
		Password: password,
	}
	hasEntry := a.storage.HasEntry(ctx, u)
	if hasEntry {
		return "", "", models.ErrNotFound
	}

	accToken, refreshToken, err := a.IssueTokensWithExpiration(ctx)

	if err != nil {
		return "", "", err
	}

	a.tokenStorage.Save(ctx, accToken) //question: in real app we need to perform some kind of transaction?
	a.tokenStorage.Save(ctx, refreshToken)
	return accToken, refreshToken, err
}

func (a *Auth) IssueTokensWithExpiration(ctx context.Context) (string, string, error) {
	accessClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccToken, err := accessToken.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString(secret)
	if err != nil {
		return "", "", err
	}
	a.tokenStorage.Save(ctx, signedAccToken)
	a.tokenStorage.Save(ctx, signedRefreshToken)
	return signedAccToken, signedRefreshToken, err
}

func (a *Auth) Verify(ctx context.Context, token string) error {
	hasEntry := a.tokenStorage.Get(ctx, token)
	if !hasEntry {
		return models.ErrNotFound
	}

	accClaims, err := a.parseClaims(token)
	if err != nil {
		return models.ErrClaimsParse
	}

	err = accClaims.Valid()
	if err != nil {
		return fmt.Errorf("invalid access_token:%w", models.ErrTokenInvalid)
	}
	return nil
}

func (a *Auth) parseClaims(token string) (*jwt.RegisteredClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{},
		func(*jwt.Token) (any, error) { return []byte(secret), nil },
	)
	if err != nil {
		return nil, err
	}
	return parsed.Claims.(*jwt.RegisteredClaims), nil
}

func (a *Auth) Deactivate(ctx context.Context, token string) {
	a.tokenStorage.Delete(ctx, token)
}
