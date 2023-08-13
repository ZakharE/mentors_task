package models

import "errors"

type User struct {
	Username string
	Password string
}

var ErrNotFound = errors.New("no such user")
