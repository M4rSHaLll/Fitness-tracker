package service

import (
	"Fitness-tracker/internal/repository/memory"
	"errors"
	"testing"
)

func TestCreateUserRequiresUsername(t *testing.T) {
	service := NewUserService(memory.NewUserRepository())

	_, err := service.CreateUser(123, "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpdateProfileRequiresExistingUser(t *testing.T) {
	service := NewUserService(memory.NewUserRepository())

	err := service.UpdateProfile(1, 80, 180, 30)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
