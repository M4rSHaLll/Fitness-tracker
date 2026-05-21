package service

import (
	"Fitness-tracker/internal/model"
	"fmt"

	"Fitness-tracker/internal/repository"
	"errors"
)

type WorkoutService struct {
	repo repository.WorkoutRepository
}

func NewWorkoutService(repo repository.WorkoutRepository) *WorkoutService {
	return &WorkoutService{
		repo: repo,
	}
}

func (s *WorkoutService) CreateWorkout(userID int64, description string) (*model.Workout, error) {
	if description == "" {
		return nil, errors.New("description is empty")
	}

	workout := model.NewWorkout(userID, description)

	if err := s.repo.Create(workout); err != nil {
		return nil, fmt.Errorf("can't create workout: %w", err)
	}

	return workout, nil
}

func (s *WorkoutService) DeleteWorkout(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("can't delete workout: %w", err)
	}
	return nil
}

func (s *WorkoutService) GetUserWorkouts(userID int64) ([]model.Workout, error) {
	workouts, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("can't get user workouts: %w", err)
	}

	return workouts, nil
}
