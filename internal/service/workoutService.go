package service

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"errors"
	"fmt"
)

type WorkoutService struct {
	repo     repository.WorkoutRepository
	userRepo repository.UserRepository
}

func NewWorkoutService(repo repository.WorkoutRepository, userRepo repository.UserRepository) *WorkoutService {
	return &WorkoutService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *WorkoutService) CreateWorkout(userID int64, description string) (*model.Workout, error) {
	if description == "" {
		return nil, fmt.Errorf("%w: description is empty", ErrInvalidInput)
	}
	if _, err := s.userRepo.GetByID(userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("can't get user: %w", err)
	}

	workout := model.NewWorkout(userID, description)

	if err := s.repo.Create(workout); err != nil {
		return nil, fmt.Errorf("can't create workout: %w", err)
	}

	return workout, nil
}

func (s *WorkoutService) DeleteWorkout(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrWorkoutNotFound
		}
		return fmt.Errorf("can't delete workout: %w", err)
	}
	return nil
}

func (s *WorkoutService) GetWorkoutByID(id int64) (*model.Workout, error) {
	workout, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, fmt.Errorf("can't get workout: %w", err)
	}
	return workout, nil
}

func (s *WorkoutService) GetUserWorkouts(userID int64) ([]*model.Workout, error) {
	if _, err := s.userRepo.GetByID(userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("can't get user: %w", err)
	}

	workouts, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("can't get user workouts: %w", err)
	}

	return workouts, nil
}
