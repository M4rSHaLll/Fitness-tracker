package service

import "errors"

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrUserNotFound    = errors.New("user not found")
	ErrWorkoutNotFound = errors.New("workout not found")
	ErrSetNotFound     = errors.New("set not found")
)
