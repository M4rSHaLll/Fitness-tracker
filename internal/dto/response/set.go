package response

import "time"

type Set struct {
	ID         int64     `json:"id"`
	ExerciseID int64     `json:"exercise_id"`
	WorkoutID  int64     `json:"workout_id"`
	Reps       int64     `json:"reps"`
	Weight     float64   `json:"weight"`
	RPE        float64   `json:"rpe"`
	CreatedAt  time.Time `json:"created_at"`
}
