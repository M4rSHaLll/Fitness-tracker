package service

import (
	"Fitness-tracker/internal/repository"
	"Fitness-tracker/internal/validation"
	"errors"
	"fmt"
)

const WeightIncrement = 2.5

type StatsService struct {
	workoutRepo repository.WorkoutRepository
	setRepo     repository.SetRepository
}

type UserStats struct {
	TotalWorkouts int64   `json:"total_workouts"`
	TotalVolume   float64 `json:"total_volume"`
	AverageRPE    float64 `json:"average_rpe"`
}

func NewStatsService(setRepo repository.SetRepository, workoutRepo repository.WorkoutRepository) *StatsService {
	return &StatsService{
		workoutRepo: workoutRepo,
		setRepo:     setRepo,
	}
}

func (s *StatsService) GetUserStats(userID int64) (*UserStats, error) {
	workouts, err := s.workoutRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve workouts: %w", err)
	}
	var totalVolume float64
	var totalRPE float64
	var totalSets int64

	for _, workout := range workouts {
		sets, err := s.setRepo.GetByWorkoutID(workout.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve sets: %w", err)
		}
		for _, set := range sets {
			totalVolume += set.Weight * float64(set.Reps)
			totalRPE += set.RPE
			totalSets++
		}
	}

	averageRPE := 0.0
	if totalSets > 0 {
		averageRPE = totalRPE / float64(totalSets)
	}

	return &UserStats{
		TotalWorkouts: int64(len(workouts)),
		TotalVolume:   totalVolume,
		AverageRPE:    averageRPE,
	}, nil
}

func (s *StatsService) CalculateRepMax(weight float64, reps int64, rpe float64) (float64, error) {
	if err := validation.ValidationReps(reps); err != nil {
		return 0, err
	}
	if err := validation.ValidationExerciseWeight(weight); err != nil {
		return 0, err
	}
	if err := validation.ValidationRPE(rpe); err != nil {
		return 0, err
	}

	correctWeight, correctReps, err := adjustTrainingLoad(weight, reps, rpe)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate rep max: %w", err)
	}

	repMax := correctWeight * (1 + float64(correctReps)/30)

	return repMax, nil

}

func adjustTrainingLoad(weight float64, reps int64, rpe float64) (float64, int64, error) {
	iRpe := int64(rpe * 10)

	switch iRpe % 10 {
	case 5:
		return weight + WeightIncrement, reps + (10 - iRpe/10), nil
	case 0:
		return weight, reps + (10 - iRpe/10), nil
	}

	return 0, 0, errors.New("invalid RPE value, must end with .0 or .5")
}
