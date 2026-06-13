package service

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"errors"
	"fmt"
	"strings"
)

type ExerciseService struct {
	repo repository.ExerciseRepository
}

func NewExerciseService(repo repository.ExerciseRepository) *ExerciseService {
	return &ExerciseService{
		repo: repo,
	}
}

func (s *ExerciseService) CreateExercise(name string) (*model.Exercise, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: exercise name is required", ErrInvalidInput)
	}

	exercise := model.NewExercise(name)
	if err := s.repo.Create(exercise); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("can't create exercise: %w", err)
	}

	return exercise, nil
}

func (s *ExerciseService) GetExerciseByID(id int64) (*model.Exercise, error) {
	exercise, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrExerciseNotFound
		}
		return nil, fmt.Errorf("can't get exercise: %w", err)
	}

	return exercise, nil
}

func (s *ExerciseService) GetExercises() ([]*model.Exercise, error) {
	exercises, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("can't get exercises: %w", err)
	}

	return exercises, nil
}

func (s *ExerciseService) DeleteExercise(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrExerciseNotFound
		}
		if errors.Is(err, repository.ErrResourceInUse) {
			return ErrResourceInUse
		}
		return fmt.Errorf("can't delete exercise: %w", err)
	}

	return nil
}
