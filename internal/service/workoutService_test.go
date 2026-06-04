package service

import (
	"Fitness-tracker/internal/repository/memory"
	"errors"
	"testing"
)

func TestCreateWorkoutRequiresExistingUser(t *testing.T) {
	userRepo := memory.NewUserRepository()
	workoutRepo := memory.NewWorkoutRepository()
	service := NewWorkoutService(workoutRepo, userRepo)

	_, err := service.CreateWorkout(1, "Leg day")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestCreateWorkoutForExistingUser(t *testing.T) {
	userRepo := memory.NewUserRepository()
	workoutRepo := memory.NewWorkoutRepository()
	userService := NewUserService(userRepo)
	workoutService := NewWorkoutService(workoutRepo, userRepo)

	user, err := userService.CreateUser(123, "art")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	workout, err := workoutService.CreateWorkout(user.ID, "Leg day")
	if err != nil {
		t.Fatalf("create workout: %v", err)
	}
	if workout.UserID != user.ID {
		t.Fatalf("expected workout user_id %d, got %d", user.ID, workout.UserID)
	}
}
