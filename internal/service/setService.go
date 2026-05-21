package service

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"Fitness-tracker/internal/validation"
	"fmt"
)

type SetService struct {
	repo repository.SetRepository
}

func NewSetService(repo repository.SetRepository) *SetService {
	return &SetService{
		repo: repo,
	}
}

func (s *SetService) CreateSet(exerciseID, workoutID int64, reps int64, weight, rpe float64) (*model.Set, error) {
	if err := validation.ValidationReps(reps); err != nil {
		return nil, err
	}
	if err := validation.ValidationExerciseWeight(weight); err != nil {
		return nil, err
	}
	if err := validation.ValidationRPE(rpe); err != nil {
		return nil, err
	}

	// TODO check if workout exists and belongs to the user

	set := model.NewSet(exerciseID, workoutID, reps, weight, rpe)

	if err := s.repo.Create(set); err != nil {
		return nil, fmt.Errorf("can't create set: %w", err)
	}

	return set, nil
}

func (s *SetService) DeleteSet(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("can't delete set: %w", err)
	}
	return nil
}

func (s *SetService) GetWorkoutSets(workoutID int64) ([]model.Set, error) {
	sets, err := s.repo.GetByWorkoutID(workoutID)
	if err != nil {
		return nil, fmt.Errorf("can't get sets: %w", err)
	}
	return sets, nil
}
