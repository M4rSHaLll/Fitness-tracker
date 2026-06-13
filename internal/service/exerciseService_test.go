package service

import (
	"Fitness-tracker/internal/repository/memory"
	"errors"
	"testing"
)

func TestCreateExerciseRequiresName(t *testing.T) {
	service := NewExerciseService(memory.NewExerciseRepository())

	_, err := service.CreateExercise("")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateExerciseRejectsBlankName(t *testing.T) {
	service := NewExerciseService(memory.NewExerciseRepository())

	_, err := service.CreateExercise("   ")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateExerciseRejectsDuplicateName(t *testing.T) {
	service := NewExerciseService(memory.NewExerciseRepository())

	if _, err := service.CreateExercise("Bench Press"); err != nil {
		t.Fatalf("create exercise: %v", err)
	}

	_, err := service.CreateExercise("bench press")
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestGetExerciseByIDRequiresExistingExercise(t *testing.T) {
	service := NewExerciseService(memory.NewExerciseRepository())

	_, err := service.GetExerciseByID(1)
	if !errors.Is(err, ErrExerciseNotFound) {
		t.Fatalf("expected ErrExerciseNotFound, got %v", err)
	}
}
