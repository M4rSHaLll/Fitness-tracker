package model

import "time"

type Set struct {
	ID         int64     `json:"id"`
	ExerciseID int64     `json:"exercise_id"`
	WorkoutID  int64     `json:"workout_id"`
	CreatedAt  time.Time `json:"created_at"`

	Reps   int64   `json:"reps"`
	Weight float64 `json:"weight"`
	RPE    float64 `json:"rpe"`
}

func NewSet(exerciseID, workoutID int64, reps int64, weight, rpe float64) *Set {
	return &Set{
		ExerciseID: exerciseID,
		WorkoutID:  workoutID,
		Reps:       reps,
		Weight:     weight,
		RPE:        rpe,
		CreatedAt:  time.Now(),
	}
}
