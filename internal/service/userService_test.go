package service

import (
	"Fitness-tracker/internal/repository/memory"
	"errors"
	"testing"
	"time"
)

func TestCreateUserRequiresUsername(t *testing.T) {
	service := NewUserService(memory.NewUserRepository())

	_, err := service.CreateUser(123, "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateUserRejectsDuplicateTelegramID(t *testing.T) {
	service := NewUserService(memory.NewUserRepository())

	if _, err := service.CreateUser(123, "art"); err != nil {
		t.Fatalf("create user: %v", err)
	}

	_, err := service.CreateUser(123, "another")
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestUpdateProfileRequiresExistingUser(t *testing.T) {
	service := NewUserService(memory.NewUserRepository())

	err := service.UpdateProfile(1, 80, 180, 30)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUpdateProfileRefreshesUpdatedAt(t *testing.T) {
	service := NewUserService(memory.NewUserRepository())

	user, err := service.CreateUser(123, "art")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	createdUpdatedAt := user.UpdatedAt

	time.Sleep(time.Millisecond)
	if err := service.UpdateProfile(user.ID, 80, 180, 30); err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if !user.UpdatedAt.After(createdUpdatedAt) {
		t.Fatalf("expected updated_at to be refreshed")
	}
}
