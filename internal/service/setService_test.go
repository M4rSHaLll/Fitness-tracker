package service

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository/memory"
	"errors"
	"testing"
)

func TestCreateSetRequiresExistingWorkout(t *testing.T) {
	workoutRepo := memory.NewWorkoutRepository()
	exerciseRepo := memory.NewExerciseRepository()
	setRepo := memory.NewSetRepository()
	service := NewSetService(setRepo, workoutRepo, exerciseRepo)

	_, err := service.CreateSet(1, 1, 5, 100, 8)
	if !errors.Is(err, ErrWorkoutNotFound) {
		t.Fatalf("expected ErrWorkoutNotFound, got %v", err)
	}
}

func TestCreateSetRequiresExistingExercise(t *testing.T) {
	workoutRepo := memory.NewWorkoutRepository()
	exerciseRepo := memory.NewExerciseRepository()
	setRepo := memory.NewSetRepository()
	service := NewSetService(setRepo, workoutRepo, exerciseRepo)
	workout := model.NewWorkout(1, "Bench press")

	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("create workout: %v", err)
	}

	_, err := service.CreateSet(1, workout.ID, 5, 100, 8)
	if !errors.Is(err, ErrExerciseNotFound) {
		t.Fatalf("expected ErrExerciseNotFound, got %v", err)
	}
}

func TestCreateSetForExistingWorkoutAndExercise(t *testing.T) {
	workoutRepo := memory.NewWorkoutRepository()
	exerciseRepo := memory.NewExerciseRepository()
	setRepo := memory.NewSetRepository()
	service := NewSetService(setRepo, workoutRepo, exerciseRepo)
	workout := model.NewWorkout(1, "Bench press")
	exercise := model.NewExercise("Bench press")

	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("create workout: %v", err)
	}
	if err := exerciseRepo.Create(exercise); err != nil {
		t.Fatalf("create exercise: %v", err)
	}

	set, err := service.CreateSet(exercise.ID, workout.ID, 5, 100, 8)
	if err != nil {
		t.Fatalf("create set: %v", err)
	}
	if set.WorkoutID != workout.ID {
		t.Fatalf("expected set workout_id %d, got %d", workout.ID, set.WorkoutID)
	}
	if set.ExerciseID != exercise.ID {
		t.Fatalf("expected set exercise_id %d, got %d", exercise.ID, set.ExerciseID)
	}
}
