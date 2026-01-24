package system

import "errors"

var (
	ErrUserExists    = errors.New("user already exists")
	ErrUserNotFound  = errors.New("user not found")
	ErrBadCredential = errors.New("invalid credentials")
)
