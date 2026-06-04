package service

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository/memory"
	"errors"
	"testing"
)

func TestCreateSetRequiresExistingWorkout(t *testing.T) {
	workoutRepo := memory.NewWorkoutRepository()
	setRepo := memory.NewSetRepository()
	service := NewSetService(setRepo, workoutRepo)

	_, err := service.CreateSet(1, 1, 5, 100, 8)
	if !errors.Is(err, ErrWorkoutNotFound) {
		t.Fatalf("expected ErrWorkoutNotFound, got %v", err)
	}
}

func TestCreateSetForExistingWorkout(t *testing.T) {
	workoutRepo := memory.NewWorkoutRepository()
	setRepo := memory.NewSetRepository()
	service := NewSetService(setRepo, workoutRepo)
	workout := model.NewWorkout(1, "Bench press")

	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("create workout: %v", err)
	}

	set, err := service.CreateSet(1, workout.ID, 5, 100, 8)
	if err != nil {
		t.Fatalf("create set: %v", err)
	}
	if set.WorkoutID != workout.ID {
		t.Fatalf("expected set workout_id %d, got %d", workout.ID, set.WorkoutID)
	}
}
