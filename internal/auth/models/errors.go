package models

import "errors"

var ErrNotFound = errors.New("no such user")
var ErrClaimsParse = errors.New("unable parse claims")
var ErrTokenInvalid = errors.New("token is expired")
